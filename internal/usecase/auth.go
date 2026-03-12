package usecase

import (
	"context"
	"errors"
	"shc/domain"
	"shc/internal/service"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MainRepo interface {
	CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

type UserRepo interface {
	GetByUUID(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error)
	GetByLogin(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error)
	UpdateUserPasswordByUUID(ctx context.Context, uuid, password string, tx *pgx.Tx) error
	Create(ctx context.Context, user domain.CreateUser, tx *pgx.Tx) error
	CreateClient(ctx context.Context, client domain.CreateClient, tx *pgx.Tx) error
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

type OtherSystemLogin interface {
	Login(ctx context.Context, args ...any) (userUUID string, userInfo map[any]any, err error)
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

func (a *AuthUseCase) LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error) {
	if a.otherSystemLogin == nil {
		return nil, errors.New("other system login is not configured")
	}

	userUUID, userInfo, err := a.otherSystemLogin.Login(ctx, args...)
	if err != nil {
		return nil, err
	}

	tx, err := a.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	pswd, err := a.pswdService.CreatePasswordHash(uuid.NewString())
	if err != nil {
		return nil, err
	}

	user, err := a.userRepo.GetByUUID(ctx, userUUID, &tx)

	if errors.Is(err, pgx.ErrNoRows) {
		createUser := domain.CreateUser{
			Login:    "client_" + uuid.NewString(),
			Password: pswd,
			Role:     domain.UserRoleClient,
		}

		if err := a.userRepo.Create(ctx, createUser, &tx); err != nil {
			return nil, err
		}

		createClient := domain.CreateClient{
			UserUUID: userUUID,
			Info:     userInfo,
		}

		if err := a.userRepo.CreateClient(ctx, createClient, &tx); err != nil {
			return nil, err
		}

		user = &domain.User{
			UUID:     userUUID,
			Login:    createUser.Login,
			Password: pswd,
			Role:     domain.UserRoleClient,
		}
	} else if err != nil {
		return nil, err
	} else {
		if err := a.userRepo.UpdateUserPasswordByUUID(ctx, user.UUID, pswd, &tx); err != nil {
			return nil, err
		}
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
