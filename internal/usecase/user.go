package usecase

import (
	"context"
	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type UserRepo interface {
	GetAll(ctx context.Context, tx *pgx.Tx) ([]domain.User, error)
	GetByUUID(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error)
	GetByLogin(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error)
	Create(ctx context.Context, user domain.CreateUser, tx *pgx.Tx) error
	UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser, tx *pgx.Tx) error
	DeleteByUUID(ctx context.Context, uuid string, tx *pgx.Tx) error
}

type UserPswdService interface {
	CreatePasswordHash(password string) (string, error)
}

type UserUseCase struct {
	mainRepo    MainRepo
	userRepo    UserRepo
	pswdService UserPswdService
}

func NewUserUseCase(mainRepo MainRepo, userRepo UserRepo, pswdService UserPswdService) *UserUseCase {
	return &UserUseCase{
		mainRepo:    mainRepo,
		userRepo:    userRepo,
		pswdService: pswdService,
	}
}

func (u *UserUseCase) GetAll(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.GetAll(ctx, nil)
}

func (u *UserUseCase) GetByUUID(ctx context.Context, uuid string) (*domain.User, error) {
	return u.userRepo.GetByUUID(ctx, uuid, nil)
}

func (u *UserUseCase) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	return u.userRepo.GetByLogin(ctx, login, nil)
}

func (u *UserUseCase) Create(ctx context.Context, user domain.CreateUser) error {
	if user.Password != "" {
		passwordHash, err := u.pswdService.CreatePasswordHash(user.Password)
		if err != nil {
			return err
		}
		user.Password = passwordHash
	}

	return u.userRepo.Create(ctx, user, nil)
}

func (u *UserUseCase) UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser) error {
	if user.Password != "" {
		passwordHash, err := u.pswdService.CreatePasswordHash(user.Password)
		if err != nil {
			return err
		}
		user.Password = passwordHash
	}

	return u.userRepo.UpdateByUUID(ctx, uuid, user, nil)
}

func (u *UserUseCase) DeleteByUUID(ctx context.Context, uuid string) error {
	return u.userRepo.DeleteByUUID(ctx, uuid, nil)
}
