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
func (u *UserUseCase) GetAll(ctx context.Context, page, limit int) ([]domain.User, error) {
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
func (u *UserUseCase) Create(ctx context.Context, user domain.CreateUser) error {
	if len(user.Password) < 6 {
		return domain.ErrShortPassword
	}

	passwordHash, err := u.pswdService.CreatePasswordHash(user.Password)
	if err != nil {
		return err
	}
	user.Password = passwordHash

	return u.userRepo.Create(ctx, user, nil)
}

// UpdateByUUID updates a user and hashes a new password when provided.
func (u *UserUseCase) UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser) error {
	if user.Password != "" {
		if len(user.Password) < 6 {
			return domain.ErrShortPassword
		}

		passwordHash, err := u.pswdService.CreatePasswordHash(user.Password)
		if err != nil {
			return err
		}
		user.Password = passwordHash
	}

	return u.userRepo.UpdateByUUID(ctx, uuid, user, nil)
}

// DeleteByUUID removes a user by UUID.
func (u *UserUseCase) DeleteByUUID(ctx context.Context, uuid string) error {
	return u.userRepo.DeleteByUUID(ctx, uuid, nil)
}
