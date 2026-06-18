package app

import (
	"context"
	"errors"
	"testing"

	"shc/config"
	"shc/domain"
	"shc/internal/repository"
	"shc/internal/service"
)

type authUseCaseStub struct {
	loginFn                func(ctx context.Context, login, password string) (*domain.JWTTokens, error)
	loginClientFn          func(ctx context.Context, args ...any) (*domain.JWTTokens, error)
	logoutFn               func(ctx context.Context, accessToken string) error
	getAccessTokenClaimsFn func(accessToken string) (*service.AccessTokenClaims, error)
	refreshFn              func(ctx context.Context, refreshToken string) (*domain.JWTTokens, error)
}

func (a authUseCaseStub) Login(ctx context.Context, login, password string) (*domain.JWTTokens, error) {
	return a.loginFn(ctx, login, password)
}

func (a authUseCaseStub) LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error) {
	return a.loginClientFn(ctx, args...)
}

func (a authUseCaseStub) Logout(ctx context.Context, accessToken string) error {
	return a.logoutFn(ctx, accessToken)
}

func (a authUseCaseStub) GetAccessTokenClaims(accessToken string) (*service.AccessTokenClaims, error) {
	return a.getAccessTokenClaimsFn(accessToken)
}

func (a authUseCaseStub) Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error) {
	return a.refreshFn(ctx, refreshToken)
}

func Test_NewApp(t *testing.T) {
	origNewRepository := newRepository
	t.Cleanup(func() {
		newRepository = origNewRepository
	})

	wantErr := errors.New("new repo error")
	newRepository = func(context.Context, string) (*repository.Repository, error) {
		return nil, wantErr
	}

	app, err := NewApp(context.Background(), config.Config{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected new repo error, got=%v", err)
	}
	if app != nil {
		t.Fatalf("expected nil app, got=%v", app)
	}

	fakeRepo := &repository.Repository{}
	newRepository = func(context.Context, string) (*repository.Repository, error) {
		return fakeRepo, nil
	}

	app, err = NewApp(context.Background(), config.Config{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if app.Repository != fakeRepo {
		t.Fatal("unexpected repository")
	}
	if app.MsgCache == nil || app.JWTService == nil || app.EncryptionService == nil {
		t.Fatal("expected app services to be initialized")
	}
	if app.AuthUseCase == nil || app.UserUseCase == nil || app.ScheduleUseCase == nil || app.DepartmentUseCase == nil || app.ChatUseCase == nil {
		t.Fatal("expected use cases to be initialized")
	}
	if app.API == nil {
		t.Fatal("expected api to be initialized")
	}
}

func Test_AuthHTTPAdapter(t *testing.T) {
	tokens := &domain.JWTTokens{AccessToken: "access", RefreshToken: "refresh"}
	wantErr := errors.New("boom")
	adapter := authHTTPAdapter{
		auth: authUseCaseStub{
			loginFn: func(_ context.Context, login, password string) (*domain.JWTTokens, error) {
				if login != "login" || password != "password" {
					t.Fatalf("unexpected credentials: %q %q", login, password)
				}
				return tokens, nil
			},
			loginClientFn: func(_ context.Context, args ...any) (*domain.JWTTokens, error) {
				if len(args) != 1 || args[0] != "arg" {
					t.Fatalf("unexpected login client args: %v", args)
				}
				return tokens, nil
			},
			logoutFn: func(_ context.Context, accessToken string) error {
				if accessToken != "access" {
					t.Fatalf("unexpected access token: %q", accessToken)
				}
				return wantErr
			},
			getAccessTokenClaimsFn: func(accessToken string) (*service.AccessTokenClaims, error) {
				if accessToken != "access" {
					t.Fatalf("unexpected access token: %q", accessToken)
				}
				return nil, wantErr
			},
			refreshFn: func(_ context.Context, refreshToken string) (*domain.JWTTokens, error) {
				if refreshToken != "refresh" {
					t.Fatalf("unexpected refresh token: %q", refreshToken)
				}
				return tokens, nil
			},
		},
	}

	loginTokens, err := adapter.Login(context.Background(), "login", "password")
	if err != nil || loginTokens != tokens {
		t.Fatalf("unexpected login result: tokens=%v err=%v", loginTokens, err)
	}

	loginClientTokens, err := adapter.LoginClient(context.Background(), "arg")
	if err != nil || loginClientTokens != tokens {
		t.Fatalf("unexpected login client result: tokens=%v err=%v", loginClientTokens, err)
	}

	if !errors.Is(adapter.Logout(context.Background(), "access"), wantErr) {
		t.Fatal("expected logout error")
	}
	if !errors.Is(adapter.GetAccessTokenClaims(context.Background(), "access"), wantErr) {
		t.Fatal("expected get access token claims error")
	}

	refreshTokens, err := adapter.Refresh(context.Background(), "refresh")
	if err != nil || refreshTokens != tokens {
		t.Fatalf("unexpected refresh result: tokens=%v err=%v", refreshTokens, err)
	}
}

func Test_Stubs(t *testing.T) {
	s3 := stubS3{}
	if _, err := s3.Save(context.Background(), "path", []byte("data")); !errors.Is(err, domain.ErrUnknownObject) {
		t.Fatalf("expected unknown object from Save, got=%v", err)
	}
	if err := s3.Delete(context.Background(), "path"); !errors.Is(err, domain.ErrUnknownObject) {
		t.Fatalf("expected unknown object from Delete, got=%v", err)
	}

	other := stubOtherSystemLogin{}
	if _, _, err := other.Login(context.Background(), "arg"); !errors.Is(err, domain.ErrUnknownObject) {
		t.Fatalf("expected unknown object from Login, got=%v", err)
	}
}
