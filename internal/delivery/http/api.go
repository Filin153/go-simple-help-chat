package http

import (
	"context"
	"net/http"
	"shc/config"
	"shc/domain"
	"shc/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
)

type AuthInterface interface {
	Login(ctx context.Context, login, password string) (*domain.JWTTokens, error)
	LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error)
	Logout(ctx context.Context, accessToken string) error
	GetAccessTokenClaims(ctx context.Context, accessToken string) (*service.AccessTokenClaims, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error)
}

type API struct {
	router    *chi.Mux
	server    *http.Server
	auth      AuthInterface
	config    config.HTTPConfig
	validator *validator.Validate
}

func NewAPI(auth AuthInterface, httpConfig config.HTTPConfig) *API {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   httpConfig.CORS.AllowedOrigins,
		AllowedMethods:   httpConfig.CORS.AllowedMethods,
		AllowedHeaders:   httpConfig.CORS.AllowedHeaders,
		ExposedHeaders:   httpConfig.CORS.ExposedHeaders,
		AllowCredentials: httpConfig.CORS.AllowCredentials,
		MaxAge:           httpConfig.CORS.MaxAge,
	}))

	api := &API{
		router: r,
		server: &http.Server{
			Addr:    httpConfig.Addr,
			Handler: r,
		},
		config:    httpConfig,
		validator: validator.New(validator.WithRequiredStructEnabled()),
		auth:      auth,
	}
	api.setup()

	return api
}

func (a *API) Run() error {
	return a.server.ListenAndServe()
}

func (a *API) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *API) setup() {
	a.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", a.userLoginHTTP)
			r.Post("/login/client", a.clientLoginHTTP)
			r.Post("/logout", a.logoutHTTP)
			r.Post("/refresh", a.refreshTokenHTTP)
		})
	})
}
