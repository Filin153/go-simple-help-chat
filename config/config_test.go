package config

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"shc/domain"
)

func Test_Default(t *testing.T) {
	cfg := Default()

	if cfg.JWT.Issuer != "simple-help-chat" {
		t.Fatalf("unexpected issuer: %q", cfg.JWT.Issuer)
	}
	if cfg.JWT.SignKey != "simple-help-chat-dev-jwt-sign-key" {
		t.Fatalf("unexpected sign key: %q", cfg.JWT.SignKey)
	}
	if cfg.JWT.AccessTokenTTL != 10*time.Minute {
		t.Fatalf("unexpected access ttl: %v", cfg.JWT.AccessTokenTTL)
	}
	if cfg.JWT.RefreshTokenTTL != 30*24*time.Hour {
		t.Fatalf("unexpected refresh ttl: %v", cfg.JWT.RefreshTokenTTL)
	}
	if string(cfg.Encryption.Key32[:]) != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("unexpected encryption key: %q", string(cfg.Encryption.Key32[:]))
	}
	if cfg.Auth.RoleScopes[domain.UserRoleAdmin][0] != "*" {
		t.Fatalf("unexpected admin scope: %v", cfg.Auth.RoleScopes[domain.UserRoleAdmin])
	}
	if cfg.HTTP.Cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected same site: %v", cfg.HTTP.Cookie.SameSite)
	}
}

func Test_ApplyDefaults(t *testing.T) {
	var key [32]byte
	copy(key[:], []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))

	roleScopes := map[domain.UserRole][]string{
		domain.UserRoleManager: {"x", "y"},
	}

	cfg := ApplyDefaults(Config{
		PostgresDSN: "postgres://dsn",
		Auth: AuthConfig{
			RoleScopes: roleScopes,
		},
		JWT: JWTConfig{
			Issuer:          "issuer",
			SignKey:         "sign-key",
			AccessTokenTTL:  time.Minute,
			RefreshTokenTTL: 2 * time.Minute,
		},
		Encryption: EncryptionConfig{
			Key32: key,
		},
		Chat: ChatConfig{
			ReadTimeout:  5 * time.Second,
			PollInterval: 2 * time.Second,
		},
		HTTP: HTTPConfig{
			Addr: ":9090",
			CORS: CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowedMethods:   []string{"PATCH"},
				AllowedHeaders:   []string{"X-Test"},
				ExposedHeaders:   []string{"X-Expose"},
				AllowCredentials: false,
				MaxAge:           1,
			},
			Cookie: CookieConfig{
				AccessTokenTTL:  3 * time.Minute,
				RefreshTokenTTL: 4 * time.Minute,
				Secure:          false,
				SameSite:        http.SameSiteLaxMode,
			},
		},
	})

	if cfg.PostgresDSN != "postgres://dsn" {
		t.Fatalf("unexpected dsn: %q", cfg.PostgresDSN)
	}
	if !reflect.DeepEqual(cfg.Auth.RoleScopes, roleScopes) {
		t.Fatalf("unexpected role scopes: %v", cfg.Auth.RoleScopes)
	}
	if cfg.JWT.Issuer != "issuer" || cfg.JWT.SignKey != "sign-key" {
		t.Fatalf("unexpected jwt config: %+v", cfg.JWT)
	}
	if cfg.JWT.AccessTokenTTL != time.Minute || cfg.JWT.RefreshTokenTTL != 2*time.Minute {
		t.Fatalf("unexpected ttl config: %+v", cfg.JWT)
	}
	if cfg.Encryption.Key32 != key {
		t.Fatal("unexpected encryption key")
	}
	if cfg.Chat.ReadTimeout != 5*time.Second || cfg.Chat.PollInterval != 2*time.Second {
		t.Fatalf("unexpected chat config: %+v", cfg.Chat)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("unexpected addr: %q", cfg.HTTP.Addr)
	}
	if !reflect.DeepEqual(cfg.HTTP.CORS.AllowedOrigins, []string{"https://example.com"}) {
		t.Fatalf("unexpected cors origins: %v", cfg.HTTP.CORS.AllowedOrigins)
	}
	if cfg.HTTP.CORS.AllowCredentials {
		t.Fatalf("unexpected allow credentials: %v", cfg.HTTP.CORS.AllowCredentials)
	}
	if cfg.HTTP.Cookie.AccessTokenTTL != 3*time.Minute || cfg.HTTP.Cookie.RefreshTokenTTL != 4*time.Minute {
		t.Fatalf("unexpected cookie ttl: %+v", cfg.HTTP.Cookie)
	}
	if cfg.HTTP.Cookie.Secure {
		t.Fatalf("unexpected secure flag: %v", cfg.HTTP.Cookie.Secure)
	}
	if cfg.HTTP.Cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie same site: %v", cfg.HTTP.Cookie.SameSite)
	}
}

func Test_IsZeroConfigs(t *testing.T) {
	if !isZeroCORSConfig(CORSConfig{}) {
		t.Fatal("expected zero cors config")
	}
	if isZeroCORSConfig(CORSConfig{AllowCredentials: true}) {
		t.Fatal("expected non-zero cors config")
	}

	if !isZeroCookieConfig(CookieConfig{}) {
		t.Fatal("expected zero cookie config")
	}
	if isZeroCookieConfig(CookieConfig{Secure: true}) {
		t.Fatal("expected non-zero cookie config")
	}
}
