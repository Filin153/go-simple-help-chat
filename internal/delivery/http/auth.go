package http

import (
	"encoding/json"
	"net/http"
	"shc/config"
	"shc/domain"
	"time"
)

type loginForm struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func setTokenToCookie(w http.ResponseWriter, tokens *domain.JWTTokens, cookieConfig config.CookieConfig) {
	cookieAccessToken := &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		Expires:  time.Now().Add(cookieConfig.AccessTokenTTL),
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		SameSite: cookieConfig.SameSite,
	}

	cookieRefreshToken := &http.Cookie{
		Name:     "refresh_token",
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
