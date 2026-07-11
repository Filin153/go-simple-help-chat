package config

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"shc/domain"
)

func Test_NewConfig_FromEnvironment(t *testing.T) {
	clearConfigEnv(t)
	setTestConfigEnv(t, nil)

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}

	if cfg.PostgresDSN != "postgres://env-dsn" {
		t.Fatalf("unexpected dsn: %q", cfg.PostgresDSN)
	}
	if cfg.MiniO.URL != "http://minio:9000" ||
		cfg.MiniO.Login != "admin" ||
		cfg.MiniO.Password != "adminadmin" ||
		cfg.MiniO.ImageBucket != "images" ||
		cfg.MiniO.PrivateBucket != "private" {
		t.Fatalf("unexpected minio config: %+v", cfg.MiniO)
	}
	if !reflect.DeepEqual(cfg.Auth.RoleScopes[domain.UserRoleAdmin], []string{"*"}) {
		t.Fatalf("unexpected admin scopes: %v", cfg.Auth.RoleScopes[domain.UserRoleAdmin])
	}
	if !reflect.DeepEqual(cfg.Auth.RoleScopes[domain.UserRoleManager], []string{"chat", "report"}) {
		t.Fatalf("unexpected manager scopes: %v", cfg.Auth.RoleScopes[domain.UserRoleManager])
	}
	if !reflect.DeepEqual(cfg.Auth.RoleScopes[domain.UserRoleClient], []string{"chat"}) {
		t.Fatalf("unexpected client scopes: %v", cfg.Auth.RoleScopes[domain.UserRoleClient])
	}
	if cfg.Admin.Login != "admin" || cfg.Admin.Password != "admin" {
		t.Fatalf("unexpected admin config: %+v", cfg.Admin)
	}
	if cfg.JWT.Issuer != "env-issuer" || cfg.JWT.SignKey != "env-sign-key" {
		t.Fatalf("unexpected jwt config: %+v", cfg.JWT)
	}
	if cfg.JWT.AccessTokenTTL != 15*time.Minute || cfg.JWT.RefreshTokenTTL != 48*time.Hour {
		t.Fatalf("unexpected jwt ttl config: %+v", cfg.JWT)
	}
	if string(cfg.Encryption.Key32[:]) != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("unexpected encryption key: %q", string(cfg.Encryption.Key32[:]))
	}
	if cfg.Chat.ReadTimeout != 5*time.Second || cfg.Chat.PollInterval != 250*time.Millisecond {
		t.Fatalf("unexpected chat config: %+v", cfg.Chat)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("unexpected addr: %q", cfg.HTTP.Addr)
	}
	if !reflect.DeepEqual(cfg.HTTP.CORS.AllowedOrigins, []string{"http://localhost:3000", "https://example.com"}) {
		t.Fatalf("unexpected cors origins: %v", cfg.HTTP.CORS.AllowedOrigins)
	}
	if cfg.HTTP.CORS.AllowCredentials {
		t.Fatalf("unexpected allow credentials: %v", cfg.HTTP.CORS.AllowCredentials)
	}
	if cfg.HTTP.CORS.MaxAge != 60 {
		t.Fatalf("unexpected max age: %d", cfg.HTTP.CORS.MaxAge)
	}
	if cfg.HTTP.Cookie.AccessTokenTTL != 7*time.Minute || cfg.HTTP.Cookie.RefreshTokenTTL != 168*time.Hour {
		t.Fatalf("unexpected cookie ttl: %+v", cfg.HTTP.Cookie)
	}
	if cfg.HTTP.Cookie.Secure {
		t.Fatalf("unexpected cookie secure: %v", cfg.HTTP.Cookie.Secure)
	}
	if cfg.HTTP.Cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie same site: %v", cfg.HTTP.Cookie.SameSite)
	}
}

func Test_NewConfig_MissingRequiredValue(t *testing.T) {
	clearConfigEnv(t)
	setTestConfigEnv(t, map[string]string{
		"HTTP_ADDR": "",
	})

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "HTTP_ADDR") {
		t.Fatalf("expected HTTP_ADDR error, got %v", err)
	}
}

func Test_NewConfig_InvalidDuration(t *testing.T) {
	clearConfigEnv(t)
	setTestConfigEnv(t, map[string]string{
		"JWT_ACCESS_TOKEN_TTL": "bad",
	})

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "JWT_ACCESS_TOKEN_TTL") {
		t.Fatalf("expected JWT_ACCESS_TOKEN_TTL error, got %v", err)
	}
}

func Test_NewConfig_InvalidEncryptionKeyFile(t *testing.T) {
	clearConfigEnv(t)
	setTestConfigEnv(t, nil)

	secretDir := t.TempDir()
	keyFile := filepath.Join(secretDir, "encryption.key")
	if err := os.WriteFile(keyFile, []byte("short\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ENCRYPTION_KEY_FILE", keyFile)

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "ENCRYPTION_KEY_FILE") {
		t.Fatalf("expected ENCRYPTION_KEY_FILE error, got %v", err)
	}
}

func setTestConfigEnv(t *testing.T, overrides map[string]string) {
	t.Helper()

	secretDir := t.TempDir()
	jwtSignKeyFile := filepath.Join(secretDir, "jwt-sign.key")
	if err := os.WriteFile(jwtSignKeyFile, []byte("env-sign-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	encryptionKeyFile := filepath.Join(secretDir, "encryption.key")
	if err := os.WriteFile(encryptionKeyFile, []byte("6161616161616161616161616161616161616161616161616161616161616161\n"), 0600); err != nil {
		t.Fatal(err)
	}

	values := map[string]string{
		"POSTGRES_DSN":             "postgres://env-dsn",
		"MINIO_URL":                "http://minio:9000",
		"MINIO_ROOT_USER":          "admin",
		"MINIO_ROOT_PASSWORD":      "adminadmin",
		"MINIO_IMAGE_BUCKET":       "images",
		"MINIO_PRIVATE_BUCKET":     "private",
		"AUTH_ADMIN_SCOPES":        "*",
		"AUTH_MANAGER_SCOPES":      "chat,report",
		"AUTH_CLIENT_SCOPES":       "chat",
		"ADMIN_LOGIN":              "admin",
		"ADMIN_PASSWORD":           "admin",
		"JWT_ISSUER":               "env-issuer",
		"JWT_SIGN_KEY_FILE":        jwtSignKeyFile,
		"JWT_ACCESS_TOKEN_TTL":     "15m",
		"JWT_REFRESH_TOKEN_TTL":    "48h",
		"ENCRYPTION_KEY_FILE":      encryptionKeyFile,
		"CHAT_READ_TIMEOUT":        "5s",
		"CHAT_POLL_INTERVAL":       "250ms",
		"HTTP_ADDR":                ":9090",
		"CORS_ALLOWED_ORIGINS":     "http://localhost:3000,https://example.com",
		"CORS_ALLOWED_METHODS":     "GET,PATCH",
		"CORS_ALLOWED_HEADERS":     "Authorization,Content-Type",
		"CORS_EXPOSED_HEADERS":     "X-Request-ID",
		"CORS_ALLOW_CREDENTIALS":   "false",
		"CORS_MAX_AGE":             "60",
		"COOKIE_ACCESS_TOKEN_TTL":  "7m",
		"COOKIE_REFRESH_TOKEN_TTL": "168h",
		"COOKIE_SECURE":            "false",
		"COOKIE_SAME_SITE":         "lax",
	}

	for key, value := range overrides {
		values[key] = value
	}

	for key, value := range values {
		t.Setenv(key, value)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"POSTGRES_DSN",
		"MINIO_URL",
		"MINIO_ROOT_USER",
		"MINIO_ROOT_PASSWORD",
		"MINIO_IMAGE_BUCKET",
		"MINIO_PRIVATE_BUCKET",
		"AUTH_ADMIN_SCOPES",
		"AUTH_MANAGER_SCOPES",
		"AUTH_CLIENT_SCOPES",
		"ADMIN_LOGIN",
		"ADMIN_PASSWORD",
		"JWT_ISSUER",
		"JWT_SIGN_KEY",
		"JWT_SIGN_KEY_FILE",
		"JWT_ACCESS_TOKEN_TTL",
		"JWT_REFRESH_TOKEN_TTL",
		"ENCRYPTION_KEY",
		"ENCRYPTION_KEY_FILE",
		"CHAT_READ_TIMEOUT",
		"CHAT_POLL_INTERVAL",
		"HTTP_ADDR",
		"CORS_ALLOWED_ORIGINS",
		"CORS_ALLOWED_METHODS",
		"CORS_ALLOWED_HEADERS",
		"CORS_EXPOSED_HEADERS",
		"CORS_ALLOW_CREDENTIALS",
		"CORS_MAX_AGE",
		"COOKIE_ACCESS_TOKEN_TTL",
		"COOKIE_REFRESH_TOKEN_TTL",
		"COOKIE_SECURE",
		"COOKIE_SAME_SITE",
	} {
		t.Setenv(key, "")
	}
}
