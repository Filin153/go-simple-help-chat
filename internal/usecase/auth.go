package usecase

import (
	"context"
	"shc/domain"
	"shc/internal/service"
	"time"

	"github.com/jackc/pgx/v5"
)

type MainRepo interface {
	CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

type UserRepo interface {
	GetByUUID(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error)
	GetByLogin(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error)
}

type RefreshTokenRepo interface {
	Get(ctx context.Context, jti string, tx *pgx.Tx) (*domain.RefreshToken, error)
	Create(ctx context.Context, jti, userUUID string, tx *pgx.Tx) error
	DeleteByUserUUID(ctx context.Context, userUUID string, tx *pgx.Tx) error
	DeleteByJTI(ctx context.Context, jti string, tx *pgx.Tx) error
}

type JWTService interface {
	CreateTokens(sub string, userRole domain.UserRole, scope []string, accessTokenTTL, refreshTokenTTL time.Duration) (tokens *domain.JWTTokens, refJTI string, err error)
	VerifyAccessToken(tokenStr string) (*service.AccessTokenClaims, error)
	VerifyRefreshToken(tokenStr string) (*service.RefreshTokenClaims, error)
}

type PswdService interface {
	CreatePasswordHash(password string) (string, error)
	VerifyPassword(password, hashedPassword string) bool
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
}

func NewAuthUseCase(roleScopes RoleScopes, accessTokenTTL, refreshTokenTTL time.Duration, mainRepo MainRepo, userRepo UserRepo, refreshTokenRepo RefreshTokenRepo, jwtService JWTService, pswdService PswdService) *AuthUseCase {
	return &AuthUseCase{
		roleScopes:       roleScopes,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		mainRepo:         mainRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		pswdService:      pswdService,
	}
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

	tokens, refJTI, err := a.jwtService.CreateTokens(user.UUID, user.Role, a.roleScopes[user.Role], a.accessTokenTTL, a.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.DeleteByUserUUID(ctx, user.UUID, &tx); err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.Create(ctx, refJTI, user.UUID, &tx); err != nil {
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

	if err := a.refreshTokenRepo.DeleteByUserUUID(ctx, token.Subject, nil); err != nil {
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

	user, err := a.userRepo.GetByUUID(ctx, refFromDB.UserUUID, &tx)
	if err != nil {
		return nil, domain.ErrLogin
	}

	tokens, refJTI, err := a.jwtService.CreateTokens(user.UUID, user.Role, a.roleScopes[user.Role], a.accessTokenTTL, a.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.DeleteByUserUUID(ctx, user.UUID, &tx); err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.Create(ctx, refJTI, user.UUID, &tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return tokens, nil
}
