package http

import (
	"encoding/json"
	"net/http"
	"shc/config"
	"shc/domain"
	"time"
)

const (
	cookieAccessTokenName  = "access_token"
	cookieRefreshTokenName = "refresh_token"
)

type loginForm struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func setTokenToCookie(w http.ResponseWriter, tokens *domain.JWTTokens, cookieConfig config.CookieConfig) {
	cookieAccessToken := &http.Cookie{
		Name:     cookieAccessTokenName,
		Value:    tokens.AccessToken,
		Path:     "/",
		Expires:  time.Now().Add(cookieConfig.AccessTokenTTL),
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		SameSite: cookieConfig.SameSite,
	}

	cookieRefreshToken := &http.Cookie{
		Name:     cookieRefreshTokenName,
		Value:    tokens.RefreshToken,
		Path:     "/",
		Expires:  time.Now().Add(cookieConfig.RefreshTokenTTL),
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		SameSite: cookieConfig.SameSite,
	}

	http.SetCookie(w, cookieAccessToken)
	http.SetCookie(w, cookieRefreshToken)
}

func unsetTokenFromCookie(w http.ResponseWriter, cookieConfig config.CookieConfig) {
	cookieAccessToken := &http.Cookie{
		Name:     cookieAccessTokenName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(cookieConfig.AccessTokenTTL),
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		SameSite: cookieConfig.SameSite,
		MaxAge:   -1,
	}

	cookieRefreshToken := &http.Cookie{
		Name:     cookieRefreshTokenName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(cookieConfig.RefreshTokenTTL),
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		SameSite: cookieConfig.SameSite,
		MaxAge:   -1,
	}

	http.SetCookie(w, cookieAccessToken)
	http.SetCookie(w, cookieRefreshToken)
}

func (a *API) userLoginHTTP(w http.ResponseWriter, r *http.Request) {
	var loginReqForm loginForm

	if err := json.NewDecoder(r.Body).Decode(&loginReqForm); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()

	if err := a.validator.Struct(loginReqForm); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	tokens, err := a.auth.Login(r.Context(), loginReqForm.Login, loginReqForm.Password)
	if err != nil {
		SendBaseResponse[any](w, 401, err.Error(), true, nil)
		return
	}

	setTokenToCookie(w, tokens, a.config.Cookie)
	SendBaseResponse[any](w, 200, "OK", false, nil)
}

func (a *API) clientLoginHTTP(w http.ResponseWriter, r *http.Request) {
	var loginReqForm loginForm

	if err := json.NewDecoder(r.Body).Decode(&loginReqForm); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()

	if err := a.validator.Struct(loginReqForm); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	tokens, err := a.auth.LoginClient(r.Context(), loginReqForm.Login, loginReqForm.Password)
	if err != nil {
		SendBaseResponse[any](w, 401, err.Error(), true, nil)
		return
	}

	setTokenToCookie(w, tokens, a.config.Cookie)
	SendBaseResponse[any](w, 200, "OK", false, nil)
}

func (a *API) logoutHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(cookieRefreshTokenName)
	if err != nil {
		SendBaseResponse[any](w, 401, err.Error(), true, nil)
		return
	} else if token == nil {
		SendBaseResponse[any](w, 401, domain.ErrEmptyObject.Error(), true, nil)
		return
	}

	err = a.auth.Logout(r.Context(), token.Value)
	if err != nil {
		SendBaseResponse[any](w, 401, err.Error(), true, nil)
		return
	}

	unsetTokenFromCookie(w, a.config.Cookie)
	SendBaseResponse[any](w, 204, "OK", false, nil)
}

func (a *API) refreshTokenHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(cookieRefreshTokenName)
	if err != nil {
		SendBaseResponse[any](w, 401, err.Error(), true, nil)
		return
	} else if token == nil {
		SendBaseResponse[any](w, 401, domain.ErrEmptyObject.Error(), true, nil)
		return
	}

	tokens, err := a.auth.Refresh(r.Context(), token.Value)
	if err != nil {
		SendBaseResponse[any](w, 401, err.Error(), true, nil)
		return
	}

	setTokenToCookie(w, tokens, a.config.Cookie)
	SendBaseResponse[any](w, 200, "OK", false, nil)
}
