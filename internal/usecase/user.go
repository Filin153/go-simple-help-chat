package usecase

import (
	"context"
	"shc/domain"

	"github.com/jackc/pgx/v5"
)

// UserRepo provides user CRUD operations.
type UserRepo interface {
	GetAll(ctx context.Context, page, limit int, tx pgx.Tx) ([]domain.User, error)
	GetByUUID(ctx context.Context, uuid string, tx pgx.Tx) (*domain.User, error)
	GetByLogin(ctx context.Context, login string, tx pgx.Tx) (*domain.User, error)
	Create(ctx context.Context, user domain.CreateUser, tx pgx.Tx) error
	UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser, tx pgx.Tx) error
	DeleteByUUID(ctx context.Context, uuid string, tx pgx.Tx) error
}

// UserPswdService hashes user passwords.
type UserPswdService interface {
	CreatePasswordHash(password string) (string, error)
}

// UserUseCase handles user management operations.
type UserUseCase struct {
	userRepo    UserRepo
	pswdService UserPswdService
}

// NewUserUseCase builds a UserUseCase with required dependencies.
func NewUserUseCase(userRepo UserRepo, pswdService UserPswdService) *UserUseCase {
	return &UserUseCase{
		userRepo:    userRepo,
		pswdService: pswdService,
	}
}

// GetAll returns all users.
func (u *UserUseCase) GetAll(ctx context.Context, user domain.UserSystemInfo, page, limit int) ([]domain.User, error) {
	if user.UserRole != domain.UserRoleAdmin {
		return nil, domain.ErrAccess
	}
	return u.userRepo.GetAll(ctx, page, limit, nil)
}

// GetByUUID returns a user by UUID.
func (u *UserUseCase) GetByUUID(ctx context.Context, uuid string) (*domain.User, error) {
	return u.userRepo.GetByUUID(ctx, uuid, nil)
}

// GetByLogin returns a user by login.
func (u *UserUseCase) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	return u.userRepo.GetByLogin(ctx, login, nil)
}

// Create validates and stores a new user.
func (u *UserUseCase) Create(ctx context.Context, user domain.UserSystemInfo, userForCreate domain.CreateUser) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}

	if len(userForCreate.Password) < 6 {
		return domain.ErrShortPassword
	}

	passwordHash, err := u.pswdService.CreatePasswordHash(userForCreate.Password)
	if err != nil {
		return err
	}
	userForCreate.Password = passwordHash

	return u.userRepo.Create(ctx, userForCreate, nil)
}

// UpdateByUUID updates a user and hashes a new password when provided.
func (u *UserUseCase) UpdateByUUID(ctx context.Context, user domain.UserSystemInfo, uuid string, userForUpdate domain.UpdateUser) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}

	if userForUpdate.Password != "" {
		if len(userForUpdate.Password) < 6 {
			return domain.ErrShortPassword
		}

		passwordHash, err := u.pswdService.CreatePasswordHash(userForUpdate.Password)
		if err != nil {
			return err
		}
		userForUpdate.Password = passwordHash
	}

	return u.userRepo.UpdateByUUID(ctx, uuid, userForUpdate, nil)
}

// DeleteByUUID removes a user by UUID.
func (u *UserUseCase) DeleteByUUID(ctx context.Context, user domain.UserSystemInfo, uuid string) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}
	return u.userRepo.DeleteByUUID(ctx, uuid, nil)
}
