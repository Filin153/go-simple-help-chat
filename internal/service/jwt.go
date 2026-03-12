package service

import (
	"fmt"
	"shc/domain"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	parseWithClaims = jwt.ParseWithClaims
	signTokenString = func(token *jwt.Token, sig []byte) (string, error) {
		return token.SignedString(sig)
	}
)

type AccessTokenClaims struct {
	UserRole domain.UserRole
	Scope    []string
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	JTI string
	jwt.RegisteredClaims
}

type JWT struct {
	issuer string
	sig    []byte
}

func NewJWT(issuer string, sig []byte) *JWT {
	return &JWT{
		issuer: issuer,
		sig:    sig,
	}
}

// return: JWTTokens(accessToken and refreshToken), JTI(refreshToken ID), error
func (j *JWT) CreateTokens(sub int, userRole domain.UserRole, scope []string, accessTokenTTL, refreshTokenTTL time.Duration) (tokens *domain.JWTTokens, refJTI string, err error) {
	var act, rft string
	act, err = j.createAccessToken(sub, userRole, scope, accessTokenTTL)
	if err != nil {
		return tokens, refJTI, err
	}

	rft, refJTI, err = j.createRefreshToken(refreshTokenTTL)
	if err != nil {
		return tokens, refJTI, err
	}

	tokens = &domain.JWTTokens{
		AccessToken:  act,
		RefreshToken: rft,
	}

	return tokens, refJTI, nil
}

func (j *JWT) VerifyAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	tok, err := parseWithClaims(
		tokenStr,
		&AccessTokenClaims{},
		func(_ *jwt.Token) (any, error) { return j.sig, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := tok.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type: %T", tok.Claims)
	}

	if !tok.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}

func (j *JWT) VerifyRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	tok, err := parseWithClaims(
		tokenStr,
		&RefreshTokenClaims{},
		func(_ *jwt.Token) (any, error) { return j.sig, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := tok.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type: %T", tok.Claims)
	}

	if !tok.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}

func (j *JWT) createAccessToken(sub int, userRole domain.UserRole, scope []string, ttl time.Duration) (string, error) {
	claims := AccessTokenClaims{
		userRole,
		scope,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   strconv.Itoa(sub),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	sigTokenString, err := signTokenString(token, j.sig)
	if err != nil {
		return "", err
	}

	return sigTokenString, nil
}

func (j *JWT) createRefreshToken(ttl time.Duration) (string, string, error) {
	jti := uuid.NewString()
	claims := RefreshTokenClaims{
		jti,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	sigTokenString, err := signTokenString(token, j.sig)
	if err != nil {
		return "", "", err
	}

	return sigTokenString, jti, nil
}
