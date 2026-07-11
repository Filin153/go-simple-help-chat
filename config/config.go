package config

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"shc/domain"
)

type Config struct {
	PostgresDSN string
	MiniO       MiniOConfig
	Auth        AuthConfig
	Admin       AdminConfig
	JWT         JWTConfig
	Encryption  EncryptionConfig
	Chat        ChatConfig
	HTTP        HTTPConfig
}

type AuthConfig struct {
	RoleScopes map[domain.UserRole][]string
}

type AdminConfig struct {
	Login    string
	Password string
}

type MiniOConfig struct {
	URL           string
	Login         string
	Password      string
	ImageBucket   string
	PrivateBucket string
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

func NewConfig() (Config, error) {
	var cfg Config

	var err error
	cfg.PostgresDSN, err = requiredEnvValue("POSTGRES_DSN")
	if err != nil {
		return Config{}, err
	}

	cfg.MiniO.URL, err = requiredEnvValue("MINIO_URL")
	if err != nil {
		return Config{}, err
	}
	cfg.MiniO.Login, err = requiredEnvValue("MINIO_ROOT_USER")
	if err != nil {
		return Config{}, err
	}
	cfg.MiniO.Password, err = requiredEnvValue("MINIO_ROOT_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	cfg.MiniO.ImageBucket, err = requiredEnvValue("MINIO_IMAGE_BUCKET")
	if err != nil {
		return Config{}, err
	}
	cfg.MiniO.PrivateBucket, err = requiredEnvValue("MINIO_PRIVATE_BUCKET")
	if err != nil {
		return Config{}, err
	}

	cfg.Auth.RoleScopes = make(map[domain.UserRole][]string, 3)
	cfg.Auth.RoleScopes[domain.UserRoleAdmin], err = requiredEnvList("AUTH_ADMIN_SCOPES")
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.RoleScopes[domain.UserRoleManager], err = requiredEnvList("AUTH_MANAGER_SCOPES")
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.RoleScopes[domain.UserRoleClient], err = requiredEnvList("AUTH_CLIENT_SCOPES")
	if err != nil {
		return Config{}, err
	}

	cfg.Admin.Login, err = requiredEnvValue("ADMIN_LOGIN")
	if err != nil {
		return Config{}, err
	}
	cfg.Admin.Password, err = requiredEnvValue("ADMIN_PASSWORD")
	if err != nil {
		return Config{}, err
	}

	cfg.JWT.Issuer, err = requiredEnvValue("JWT_ISSUER")
	if err != nil {
		return Config{}, err
	}
	signKeyFile, err := requiredEnvValue("JWT_SIGN_KEY_FILE")
	if err != nil {
		return Config{}, err
	}
	cfg.JWT.SignKey, err = readSecretTextFile(signKeyFile, "JWT_SIGN_KEY_FILE")
	if err != nil {
		return Config{}, err
	}
	cfg.JWT.AccessTokenTTL, err = requiredDuration("JWT_ACCESS_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}
	cfg.JWT.RefreshTokenTTL, err = requiredDuration("JWT_REFRESH_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}

	encryptionKeyFile, err := requiredEnvValue("ENCRYPTION_KEY_FILE")
	if err != nil {
		return Config{}, err
	}
	cfg.Encryption.Key32, err = readEncryptionKeyFile(encryptionKeyFile)
	if err != nil {
		return Config{}, err
	}

	cfg.Chat.ReadTimeout, err = requiredDuration("CHAT_READ_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	cfg.Chat.PollInterval, err = requiredDuration("CHAT_POLL_INTERVAL")
	if err != nil {
		return Config{}, err
	}

	cfg.HTTP.Addr, err = requiredEnvValue("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.CORS.AllowedOrigins, err = requiredEnvList("CORS_ALLOWED_ORIGINS")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.CORS.AllowedMethods, err = requiredEnvList("CORS_ALLOWED_METHODS")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.CORS.AllowedHeaders, err = requiredEnvList("CORS_ALLOWED_HEADERS")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.CORS.ExposedHeaders, err = requiredEnvList("CORS_EXPOSED_HEADERS")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.CORS.AllowCredentials, err = requiredBool("CORS_ALLOW_CREDENTIALS")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.CORS.MaxAge, err = requiredInt("CORS_MAX_AGE")
	if err != nil {
		return Config{}, err
	}

	cfg.HTTP.Cookie.AccessTokenTTL, err = requiredDuration("COOKIE_ACCESS_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.Cookie.RefreshTokenTTL, err = requiredDuration("COOKIE_REFRESH_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.Cookie.Secure, err = requiredBool("COOKIE_SECURE")
	if err != nil {
		return Config{}, err
	}
	sameSite, err := requiredEnvValue("COOKIE_SAME_SITE")
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.Cookie.SameSite, err = parseSameSite(sameSite)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func requiredEnvValue(key string) (string, error) {
	value := envValue(key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func envValue(key string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return ""
}

func readSecretTextFile(path, envName string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s: %w", envName, err)
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("%s: secret file is empty", envName)
	}

	return value, nil
}

func readEncryptionKeyFile(path string) ([32]byte, error) {
	var key [32]byte

	data, err := os.ReadFile(path)
	if err != nil {
		return key, fmt.Errorf("ENCRYPTION_KEY_FILE: %w", err)
	}

	if len(data) == 32 {
		copy(key[:], data)
		return key, nil
	}

	value := strings.TrimSpace(string(data))
	if decoded, err := hex.DecodeString(value); err == nil && len(decoded) == 32 {
		copy(key[:], decoded)
		return key, nil
	}
	if len(value) == 32 {
		copy(key[:], []byte(value))
		return key, nil
	}

	return key, fmt.Errorf("ENCRYPTION_KEY_FILE must contain 32 raw bytes, 32 text bytes, or 64 hex characters")
}

func requiredEnvList(key string) ([]string, error) {
	value, err := requiredEnvValue(key)
	if err != nil {
		return nil, err
	}

	values := splitEnvList(value)
	if len(values) == 0 {
		return nil, fmt.Errorf("%s must contain at least one value", key)
	}

	return values, nil
}

func splitEnvList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func requiredDuration(key string) (time.Duration, error) {
	value, err := requiredEnvValue(key)
	if err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return duration, nil
}

func requiredBool(key string) (bool, error) {
	value, err := requiredEnvValue(key)
	if err != nil {
		return false, err
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}

	return parsed, nil
}

func requiredInt(key string) (int, error) {
	value, err := requiredEnvValue(key)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return parsed, nil
}

func parseSameSite(value string) (http.SameSite, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "default":
		return http.SameSiteDefaultMode, nil
	case "lax":
		return http.SameSiteLaxMode, nil
	case "strict":
		return http.SameSiteStrictMode, nil
	case "none":
		return http.SameSiteNoneMode, nil
	default:
		return http.SameSiteDefaultMode, fmt.Errorf("COOKIE_SAME_SITE must be one of: default, lax, strict, none")
	}
}
