package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"shc/domain"

	"github.com/golang-jwt/jwt/v5"
)

func useDefaultJWTHooks(t *testing.T) {
	t.Helper()

	oldSignTokenString := signTokenString
	oldParseWithClaims := parseWithClaims
	t.Cleanup(func() {
		signTokenString = oldSignTokenString
		parseWithClaims = oldParseWithClaims
	})
}

func Test_NewJWT(t *testing.T) {
	issuer := "test-issuer"
	sig := []byte("test-signature")

	j := NewJWT(issuer, sig)
	if j == nil {
		t.Fatal("NewJWT must return non-nil JWT")
	}
	if j.issuer != issuer {
		t.Fatalf("issuer mismatch; got=%q want=%q", j.issuer, issuer)
	}
	if string(j.sig) != string(sig) {
		t.Fatalf("signature mismatch; got=%q want=%q", string(j.sig), string(sig))
	}
}

func Test_JWT_CreateTokens_OK(t *testing.T) {
	j := NewJWT("issuer", []byte("signature-key"))
	scope := []string{"tickets:read", "tickets:write"}

	tokens, jti, err := j.CreateTokens(1, domain.UserRoleManager, scope, time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("CreateTokens returned error: %v", err)
	}
	if tokens == nil {
		t.Fatal("CreateTokens returned nil tokens")
	}
	if tokens.AccessToken == "" {
		t.Fatal("CreateTokens returned empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Fatal("CreateTokens returned empty refresh token")
	}
	if jti == "" {
		t.Fatal("CreateTokens returned empty refresh JTI")
	}

	parsedAccessToken, err := j.VerifyAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken returned error: %v", err)
	}
	if parsedAccessToken == nil {
		t.Fatal("VerifyAccessToken returned nil claims")
	}
	if parsedAccessToken.Subject != "1" {
		t.Fatalf("subject mismatch; got=%q want=%q", parsedAccessToken.Subject, "1")
	}
	if parsedAccessToken.Issuer != "issuer" {
		t.Fatalf("issuer mismatch; got=%q want=%q", parsedAccessToken.Issuer, "issuer")
	}
	if parsedAccessToken.UserRole != domain.UserRoleManager {
		t.Fatalf("user role mismatch; got=%q want=%q", parsedAccessToken.UserRole, domain.UserRoleManager)
	}
	if len(parsedAccessToken.Scope) != len(scope) || parsedAccessToken.Scope[0] != scope[0] || parsedAccessToken.Scope[1] != scope[1] {
		t.Fatalf("scope mismatch; got=%v want=%v", parsedAccessToken.Scope, scope)
	}
	if parsedAccessToken.ExpiresAt == nil {
		t.Fatal("access token ExpiresAt must be set")
	}
	if parsedAccessToken.IssuedAt == nil {
		t.Fatal("access token IssuedAt must be set")
	}
	if parsedAccessToken.NotBefore == nil {
		t.Fatal("access token NotBefore must be set")
	}

	parsedRefreshToken, err := j.VerifyRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("VerifyRefreshToken returned error: %v", err)
	}
	if parsedRefreshToken == nil {
		t.Fatal("VerifyRefreshToken returned nil claims")
	}
	if parsedRefreshToken.JTI != jti {
		t.Fatalf("refresh JTI mismatch; got=%q want=%q", parsedRefreshToken.JTI, jti)
	}
	if parsedRefreshToken.Issuer != "issuer" {
		t.Fatalf("refresh issuer mismatch; got=%q want=%q", parsedRefreshToken.Issuer, "issuer")
	}
	if parsedRefreshToken.ExpiresAt == nil {
		t.Fatal("refresh token ExpiresAt must be set")
	}
	if parsedRefreshToken.IssuedAt == nil {
		t.Fatal("refresh token IssuedAt must be set")
	}
	if parsedRefreshToken.NotBefore == nil {
		t.Fatal("refresh token NotBefore must be set")
	}
}

func Test_JWT_CreateTokens_AccessSignError(t *testing.T) {
	useDefaultJWTHooks(t)

	signTokenString = func(_ *jwt.Token, _ []byte) (string, error) {
		return "", errors.New("sign access error")
	}

	j := NewJWT("issuer", []byte("signature-key"))
	tokens, jti, err := j.CreateTokens(1, domain.UserRoleAdmin, []string{"a"}, time.Minute, time.Hour)
	if err == nil {
		t.Fatal("CreateTokens must return error when access signing fails")
	}
	if !strings.Contains(err.Error(), "sign access error") {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens != nil {
		t.Fatalf("tokens must be nil on error, got=%+v", tokens)
	}
	if jti != "" {
		t.Fatalf("jti must be empty on error, got=%q", jti)
	}
}

func Test_JWT_CreateTokens_RefreshSignError(t *testing.T) {
	useDefaultJWTHooks(t)

	defaultSign := signTokenString
	callCount := 0
	signTokenString = func(token *jwt.Token, sig []byte) (string, error) {
		callCount++
		if callCount == 2 {
			return "", errors.New("sign refresh error")
		}
		return defaultSign(token, sig)
	}

	j := NewJWT("issuer", []byte("signature-key"))
	tokens, jti, err := j.CreateTokens(1, domain.UserRoleClient, []string{"a"}, time.Minute, time.Hour)
	if err == nil {
		t.Fatal("CreateTokens must return error when refresh signing fails")
	}
	if !strings.Contains(err.Error(), "sign refresh error") {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens != nil {
		t.Fatalf("tokens must be nil on error, got=%+v", tokens)
	}
	if jti != "" {
		t.Fatalf("jti must be empty on error, got=%q", jti)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 sign calls, got=%d", callCount)
	}
}

func Test_JWT_VerifyAccessToken_ParseError(t *testing.T) {
	j := NewJWT("issuer", []byte("signature-key"))
	claims, err := j.VerifyAccessToken("not-a-token")
	if err == nil {
		t.Fatal("VerifyAccessToken must return parse error")
	}
	if claims != nil {
		t.Fatalf("claims must be nil on error, got=%+v", claims)
	}
}

func Test_JWT_VerifyRefreshToken_ParseError(t *testing.T) {
	j := NewJWT("issuer", []byte("signature-key"))
	claims, err := j.VerifyRefreshToken("not-a-token")
	if err == nil {
		t.Fatal("VerifyRefreshToken must return parse error")
	}
	if claims != nil {
		t.Fatalf("claims must be nil on error, got=%+v", claims)
	}
}

func Test_JWT_VerifyAccessToken_UnexpectedClaimsType(t *testing.T) {
	useDefaultJWTHooks(t)

	parseWithClaims = func(_ string, _ jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{
			Claims: jwt.MapClaims{},
			Valid:  true,
		}, nil
	}

	j := NewJWT("issuer", []byte("signature-key"))
	claims, err := j.VerifyAccessToken("token")
	if err == nil {
		t.Fatal("VerifyAccessToken must return unexpected claims type error")
	}
	if !strings.Contains(err.Error(), "unexpected claims type") {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims != nil {
		t.Fatalf("claims must be nil on error, got=%+v", claims)
	}
}

func Test_JWT_VerifyAccessToken_NotValid(t *testing.T) {
	useDefaultJWTHooks(t)

	parseWithClaims = func(_ string, _ jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{
			Claims: &AccessTokenClaims{},
			Valid:  false,
		}, nil
	}

	j := NewJWT("issuer", []byte("signature-key"))
	claims, err := j.VerifyAccessToken("token")
	if err == nil {
		t.Fatal("VerifyAccessToken must return error for invalid token")
	}
	if !strings.Contains(err.Error(), "token is not valid") {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims != nil {
		t.Fatalf("claims must be nil on error, got=%+v", claims)
	}
}

func Test_JWT_VerifyRefreshToken_UnexpectedClaimsType(t *testing.T) {
	useDefaultJWTHooks(t)

	parseWithClaims = func(_ string, _ jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{
			Claims: jwt.MapClaims{},
			Valid:  true,
		}, nil
	}

	j := NewJWT("issuer", []byte("signature-key"))
	claims, err := j.VerifyRefreshToken("token")
	if err == nil {
		t.Fatal("VerifyRefreshToken must return unexpected claims type error")
	}
	if !strings.Contains(err.Error(), "unexpected claims type") {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims != nil {
		t.Fatalf("claims must be nil on error, got=%+v", claims)
	}
}

func Test_JWT_VerifyRefreshToken_NotValid(t *testing.T) {
	useDefaultJWTHooks(t)

	parseWithClaims = func(_ string, _ jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{
			Claims: &RefreshTokenClaims{},
			Valid:  false,
		}, nil
	}

	j := NewJWT("issuer", []byte("signature-key"))
	claims, err := j.VerifyRefreshToken("token")
	if err != nil {
		if !strings.Contains(err.Error(), "token is not valid") {
			t.Fatalf("unexpected error: %v", err)
		}
	} else {
		t.Fatal("VerifyRefreshToken must return error for invalid token")
	}
	if claims != nil {
		t.Fatalf("claims must be nil on error, got=%+v", claims)
	}
}
