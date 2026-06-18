package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"shc/config"
	"shc/domain"
)

type authStub struct {
	loginFn                func(ctx context.Context, login, password string) (*domain.JWTTokens, error)
	loginClientFn          func(ctx context.Context, args ...any) (*domain.JWTTokens, error)
	logoutFn               func(ctx context.Context, accessToken string) error
	getAccessTokenClaimsFn func(ctx context.Context, accessToken string) error
	refreshFn              func(ctx context.Context, refreshToken string) (*domain.JWTTokens, error)
}

func (a authStub) Login(ctx context.Context, login, password string) (*domain.JWTTokens, error) {
	return a.loginFn(ctx, login, password)
}

func (a authStub) LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error) {
	return a.loginClientFn(ctx, args...)
}

func (a authStub) Logout(ctx context.Context, accessToken string) error {
	return a.logoutFn(ctx, accessToken)
}

func (a authStub) GetAccessTokenClaims(ctx context.Context, accessToken string) error {
	return a.getAccessTokenClaimsFn(ctx, accessToken)
}

func (a authStub) Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error) {
	return a.refreshFn(ctx, refreshToken)
}

type errWriter struct {
	header nethttp.Header
}

type badJSON struct{}

func (badJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}

func (e *errWriter) Header() nethttp.Header {
	if e.header == nil {
		e.header = make(nethttp.Header)
	}
	return e.header
}

func (*errWriter) WriteHeader(int) {}

func (*errWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func newTestAPI(auth AuthInterface) *API {
	return NewAPI(auth, config.HTTPConfig{
		Addr: "127.0.0.1:0",
		Cookie: config.CookieConfig{
			AccessTokenTTL:  time.Minute,
			RefreshTokenTTL: 2 * time.Minute,
			Secure:          true,
			SameSite:        nethttp.SameSiteLaxMode,
		},
	})
}

func Test_NewAPI_And_LoginFlow(t *testing.T) {
	tokens := &domain.JWTTokens{AccessToken: "access", RefreshToken: "refresh"}
	api := newTestAPI(authStub{
		loginFn: func(_ context.Context, login, password string) (*domain.JWTTokens, error) {
			if login != "login" || password != "password" {
				t.Fatalf("unexpected credentials: %q %q", login, password)
			}
			return tokens, nil
		},
		loginClientFn:          func(context.Context, ...any) (*domain.JWTTokens, error) { return nil, nil },
		logoutFn:               func(context.Context, string) error { return nil },
		getAccessTokenClaimsFn: func(context.Context, string) error { return nil },
		refreshFn:              func(context.Context, string) (*domain.JWTTokens, error) { return nil, nil },
	})

	if api.router == nil || api.server == nil || api.validator == nil {
		t.Fatal("expected api internals to be initialized")
	}
	if api.auth == nil {
		t.Fatal("expected auth to be initialized")
	}

	req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"login","password":"password"}`))
	rec := httptest.NewRecorder()
	api.router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("unexpected status: got=%d want=200", rec.Code)
	}
	if len(rec.Result().Cookies()) != 2 {
		t.Fatalf("unexpected cookies: %v", rec.Result().Cookies())
	}

	var resp baseResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Msg != "OK" || resp.Error {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func Test_UserLoginHTTP_Errors(t *testing.T) {
	api := newTestAPI(authStub{
		loginFn: func(context.Context, string, string) (*domain.JWTTokens, error) {
			return nil, errors.New("auth error")
		},
		loginClientFn:          func(context.Context, ...any) (*domain.JWTTokens, error) { return nil, nil },
		logoutFn:               func(context.Context, string) error { return nil },
		getAccessTokenClaimsFn: func(context.Context, string) error { return nil },
		refreshFn:              func(context.Context, string) (*domain.JWTTokens, error) { return nil, nil },
	})

	tests := []struct {
		name string
		body string
		code int
	}{
		{name: "invalid json", body: `{`, code: 422},
		{name: "validation error", body: `{"login":"login"}`, code: 422},
		{name: "auth error", body: `{"login":"login","password":"password"}`, code: 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			api.userLoginHTTP(rec, req)

			if rec.Code != tt.code {
				t.Fatalf("unexpected status: got=%d want=%d", rec.Code, tt.code)
			}
		})
	}
}

func Test_SendBaseResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	data := "payload"
	if err := SendBaseResponse(rec, 201, "created", false, &data); err != nil {
		t.Fatalf("SendBaseResponse returned error: %v", err)
	}
	if rec.Code != 201 {
		t.Fatalf("unexpected status: got=%d want=201", rec.Code)
	}

	var resp baseResponse[string]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Msg != "created" || resp.Error || resp.Data == nil || *resp.Data != data {
		t.Fatalf("unexpected response: %+v", resp)
	}

	if err := SendBaseResponse[any](&errWriter{}, 200, "ok", false, nil); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("expected write error, got=%v", err)
	}

	bad := badJSON{}
	if err := SendBaseResponse(httptest.NewRecorder(), 200, "ok", false, &bad); err == nil {
		t.Fatal("expected marshal error")
	}
}

func Test_API_Run_And_Stop(t *testing.T) {
	api := newTestAPI(authStub{
		loginFn:                func(context.Context, string, string) (*domain.JWTTokens, error) { return nil, nil },
		loginClientFn:          func(context.Context, ...any) (*domain.JWTTokens, error) { return nil, nil },
		logoutFn:               func(context.Context, string) error { return nil },
		getAccessTokenClaimsFn: func(context.Context, string) error { return nil },
		refreshFn:              func(context.Context, string) (*domain.JWTTokens, error) { return nil, nil },
	})

	api.server = &nethttp.Server{Addr: "127.0.0.1:invalid", Handler: api.router}
	if err := api.Run(); err == nil {
		t.Fatal("expected run error")
	}

	api.server = &nethttp.Server{Addr: "127.0.0.1:0", Handler: api.router}
	if err := api.Stop(context.Background()); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
}
