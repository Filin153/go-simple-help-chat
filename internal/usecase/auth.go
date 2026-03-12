package usecase

import (
	"context"
	"errors"
	"shc/domain"
	"shc/internal/service"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type MainRepo interface {
	CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

type UserRepo interface {
	GetByID(ctx context.Context, id int, tx *pgx.Tx) (*domain.User, error)
	GetByLogin(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error)
}

type RefreshTokenRepo interface {
	Get(ctx context.Context, jti string, tx *pgx.Tx) (*domain.RefreshToken, error)
	Create(ctx context.Context, jti string, userID int, tx *pgx.Tx) error
	DeleteByUserID(ctx context.Context, userID int, tx *pgx.Tx) error
	DeleteByJTI(ctx context.Context, jti string, tx *pgx.Tx) error
}

type JWTService interface {
	CreateTokens(sub int, userRole domain.UserRole, scope []string, accessTokenTTL, refreshTokenTTL time.Duration) (tokens *domain.JWTTokens, refJTI string, err error)
	VerifyAccessToken(tokenStr string) (*service.AccessTokenClaims, error)
	VerifyRefreshToken(tokenStr string) (*service.RefreshTokenClaims, error)
}

type PswdService interface {
	CreatePasswordHash(password string) (string, error)
	VerifyPassword(password, hashedPassword string) bool
}

type OtherSystemLogin interface {
	Login(ctx context.Context, args ...string) (userInfo map[any]any, err error)
}

type RoleScopes map[domain.UserRole][]string

type AuthUseCase struct {
	roleScopes                      RoleScopes
	accessTokenTTL, refreshTokenTTL time.Duration
	mainRepo                        MainRepo
	userRepo                        UserRepo
	refreshTokenRepo                RefreshTokenRepo
	jwtService                      JWTService
	pswdService                     PswdService
	otherSystemLogin                OtherSystemLogin
}

func NewAuthUseCase(roleScopes RoleScopes, accessTokenTTL, refreshTokenTTL time.Duration, mainRepo MainRepo, userRepo UserRepo, refreshTokenRepo RefreshTokenRepo, jwtService JWTService, pswdService PswdService, otherSystemLogin OtherSystemLogin) *AuthUseCase {
	return &AuthUseCase{
		roleScopes:       roleScopes,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		mainRepo:         mainRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		pswdService:      pswdService,
		otherSystemLogin: otherSystemLogin,
	}
}

func (a *AuthUseCase) LoginClient(ctx context.Context, args ...string) (*domain.JWTTokens, error) {
	if a.otherSystemLogin == nil {
		return nil, errors.New("other system login is not configured")
	}

	if _, err := a.otherSystemLogin.Login(ctx, args...); err != nil {
		return nil, err
	}

	return nil, errors.New("login client is not implemented")
}

func (a *AuthUseCase) Login(ctx context.Context, login, password string) (*domain.JWTTokens, error) {
	tx, err := a.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user, err := a.userRepo.GetByLogin(ctx, login, &tx)
	if err != nil {
		return nil, domain.ErrLogin
	}

	if !a.pswdService.VerifyPassword(password, user.Password) {
		return nil, domain.ErrLogin
	}

	tokens, refJTI, err := a.jwtService.CreateTokens(user.ID, user.Role, a.roleScopes[user.Role], a.accessTokenTTL, a.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.DeleteByUserID(ctx, user.ID, &tx); err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.Create(ctx, refJTI, user.ID, &tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (a *AuthUseCase) Logout(ctx context.Context, accessToken string) error {
	token, err := a.jwtService.VerifyAccessToken(accessToken)
	if err != nil {
		return err
	}

	userID, err := strconv.Atoi(token.Subject)
	if err != nil {
		return err
	}

	if err := a.refreshTokenRepo.DeleteByUserID(ctx, userID, nil); err != nil {
		return err
	}

	return nil
}

func (a *AuthUseCase) GetAccessTokenClaims(accessToken string) (*service.AccessTokenClaims, error) {
	return a.jwtService.VerifyAccessToken(accessToken)
}

func (a *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error) {
	token, err := a.jwtService.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	tx, err := a.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	refFromDB, err := a.refreshTokenRepo.Get(ctx, token.JTI, &tx)
	if err != nil {
		return nil, err
	}

	user, err := a.userRepo.GetByID(ctx, refFromDB.UserID, &tx)
	if err != nil {
		return nil, domain.ErrLogin
	}

	tokens, refJTI, err := a.jwtService.CreateTokens(user.ID, user.Role, a.roleScopes[user.Role], a.accessTokenTTL, a.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.DeleteByUserID(ctx, user.ID, &tx); err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.Create(ctx, refJTI, user.ID, &tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return tokens, nil
}
