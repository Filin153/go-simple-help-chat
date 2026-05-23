package config

import (
	"net/http"
	"time"

	"shc/domain"
)

type Config struct {
	PostgresDSN string
	Auth        AuthConfig
	JWT         JWTConfig
	Encryption  EncryptionConfig
	Chat        ChatConfig
	HTTP        HTTPConfig
}

type AuthConfig struct {
	RoleScopes map[domain.UserRole][]string
}

type JWTConfig struct {
	Issuer          string
	SignKey         string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type EncryptionConfig struct {
	Key32 [32]byte
}

type ChatConfig struct {
	ReadTimeout  time.Duration
	PollInterval time.Duration
}

type HTTPConfig struct {
	Addr   string
	CORS   CORSConfig
	Cookie CookieConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type CookieConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Secure          bool
	SameSite        http.SameSite
}

func Default() Config {
	var encryptionKey [32]byte
	copy(encryptionKey[:], []byte("0123456789abcdef0123456789abcdef"))

	return Config{
		Auth: AuthConfig{
			RoleScopes: map[domain.UserRole][]string{
				domain.UserRoleAdmin:   {"*"},
				domain.UserRoleManager: {"chat", "department", "schedule"},
				domain.UserRoleClient:  {"chat"},
			},
		},
		JWT: JWTConfig{
			Issuer:          "simple-help-chat",
			SignKey:         "simple-help-chat-dev-jwt-sign-key",
			AccessTokenTTL:  10 * time.Minute,
			RefreshTokenTTL: 30 * 24 * time.Hour,
		},
		Encryption: EncryptionConfig{
			Key32: encryptionKey,
		},
		Chat: ChatConfig{
			ReadTimeout:  30 * time.Second,
			PollInterval: time.Second,
		},
		HTTP: HTTPConfig{
			Addr: ":8080",
			CORS: CORSConfig{
				AllowedOrigins:   []string{"https://*", "http://*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
				AllowedHeaders:   []string{""},
				ExposedHeaders:   []string{""},
				AllowCredentials: true,
				MaxAge:           300,
			},
			Cookie: CookieConfig{
				AccessTokenTTL:  10 * time.Minute,
				RefreshTokenTTL: 30 * 24 * time.Hour,
				Secure:          true,
				SameSite:        http.SameSiteStrictMode,
			},
		},
	}
}

func ApplyDefaults(cfg Config) Config {
	defaults := Default()

	if cfg.PostgresDSN != "" {
		defaults.PostgresDSN = cfg.PostgresDSN
	}

	if len(cfg.Auth.RoleScopes) > 0 {
		defaults.Auth.RoleScopes = cfg.Auth.RoleScopes
	}

	if cfg.JWT.Issuer != "" {
		defaults.JWT.Issuer = cfg.JWT.Issuer
	}
	if cfg.JWT.SignKey != "" {
		defaults.JWT.SignKey = cfg.JWT.SignKey
	}
	if cfg.JWT.AccessTokenTTL != 0 {
		defaults.JWT.AccessTokenTTL = cfg.JWT.AccessTokenTTL
	}
	if cfg.JWT.RefreshTokenTTL != 0 {
		defaults.JWT.RefreshTokenTTL = cfg.JWT.RefreshTokenTTL
	}

	if cfg.Encryption.Key32 != ([32]byte{}) {
		defaults.Encryption.Key32 = cfg.Encryption.Key32
	}

	if cfg.Chat.ReadTimeout != 0 {
		defaults.Chat.ReadTimeout = cfg.Chat.ReadTimeout
	}
	if cfg.Chat.PollInterval != 0 {
		defaults.Chat.PollInterval = cfg.Chat.PollInterval
	}

	if cfg.HTTP.Addr != "" {
		defaults.HTTP.Addr = cfg.HTTP.Addr
	}
	if len(cfg.HTTP.CORS.AllowedOrigins) > 0 {
		defaults.HTTP.CORS.AllowedOrigins = cfg.HTTP.CORS.AllowedOrigins
	}
	if len(cfg.HTTP.CORS.AllowedMethods) > 0 {
		defaults.HTTP.CORS.AllowedMethods = cfg.HTTP.CORS.AllowedMethods
	}
	if len(cfg.HTTP.CORS.AllowedHeaders) > 0 {
		defaults.HTTP.CORS.AllowedHeaders = cfg.HTTP.CORS.AllowedHeaders
	}
	if len(cfg.HTTP.CORS.ExposedHeaders) > 0 {
		defaults.HTTP.CORS.ExposedHeaders = cfg.HTTP.CORS.ExposedHeaders
	}
	if cfg.HTTP.CORS.MaxAge != 0 {
		defaults.HTTP.CORS.MaxAge = cfg.HTTP.CORS.MaxAge
	}
	if !isZeroCORSConfig(cfg.HTTP.CORS) {
		defaults.HTTP.CORS.AllowCredentials = cfg.HTTP.CORS.AllowCredentials
	}

	if cfg.HTTP.Cookie.AccessTokenTTL != 0 {
		defaults.HTTP.Cookie.AccessTokenTTL = cfg.HTTP.Cookie.AccessTokenTTL
	}
	if cfg.HTTP.Cookie.RefreshTokenTTL != 0 {
		defaults.HTTP.Cookie.RefreshTokenTTL = cfg.HTTP.Cookie.RefreshTokenTTL
	}
	if cfg.HTTP.Cookie.SameSite != 0 {
		defaults.HTTP.Cookie.SameSite = cfg.HTTP.Cookie.SameSite
	}
	if !isZeroCookieConfig(cfg.HTTP.Cookie) {
		defaults.HTTP.Cookie.Secure = cfg.HTTP.Cookie.Secure
	}

	return defaults
}

func isZeroCORSConfig(cfg CORSConfig) bool {
	return len(cfg.AllowedOrigins) == 0 &&
		len(cfg.AllowedMethods) == 0 &&
		len(cfg.AllowedHeaders) == 0 &&
		len(cfg.ExposedHeaders) == 0 &&
		!cfg.AllowCredentials &&
		cfg.MaxAge == 0
}

func isZeroCookieConfig(cfg CookieConfig) bool {
	return cfg.AccessTokenTTL == 0 &&
		cfg.RefreshTokenTTL == 0 &&
		!cfg.Secure &&
		cfg.SameSite == 0
}
