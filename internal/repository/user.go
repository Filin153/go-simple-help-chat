package repository

import (
	"context"
	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	repo *Repository
}

func NewUserRepo(repo *Repository) *UserRepo {
	return &UserRepo{
		repo: repo,
	}
}

func (u *UserRepo) GetAll(ctx context.Context, page, limit int, tx pgx.Tx) ([]domain.User, error) {
	const query = `SELECT * FROM "users" LIMIT $1 OFFSET $2;`
	rows, err := u.repo.GetDB(nil).Query(ctx, query, limit, getOffset(page, limit))
	if err != nil {
		return []domain.User{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return []domain.User{}, err
	}

	return res, nil
}

func (u *UserRepo) GetByUUID(ctx context.Context, uuid string, tx pgx.Tx) (*domain.User, error) {
	const query = `SELECT * FROM "users" WHERE "uuid"=$1;`
	rows, err := u.repo.GetDB(tx).Query(ctx, query, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (u *UserRepo) GetByLogin(ctx context.Context, login string, tx pgx.Tx) (*domain.User, error) {
	const query = `SELECT * FROM "users" WHERE "login"=$1;`
	rows, err := u.repo.GetDB(tx).Query(ctx, query, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (u *UserRepo) Create(ctx context.Context, user domain.CreateUser, tx pgx.Tx) error {
	const query = `INSERT INTO "users"("login", "password", "role") VALUES ($1, $2, $3);`
	_, err := u.repo.GetDB(tx).Exec(ctx, query, user.Login, user.Password, user.Role)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserRepo) UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser, tx pgx.Tx) error {
	const query = `UPDATE "users" SET "login"=$2, "password"=$3, "role"=$4 WHERE "uuid"=$1;`
	tag, err := u.repo.GetDB(tx).Exec(ctx, query, uuid, user.Login, user.Password, user.Role)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUnknownObject
	}
	return nil
}

func (u *UserRepo) DeleteByUUID(ctx context.Context, uuid string, tx pgx.Tx) error {
	const query = `DELETE FROM "users" WHERE "uuid"=$1;`
	tag, err := u.repo.GetDB(tx).Exec(ctx, query, uuid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUnknownObject
	}
	return nil
}

func (u *UserRepo) UpdateUserPasswordByUUID(ctx context.Context, uuid, password string, tx pgx.Tx) error {
	const query = `UPDATE "users" SET "password"=$2 WHERE "uuid"=$1;`
	tag, err := u.repo.GetDB(tx).Exec(ctx, query, uuid, password)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUnknownObject
	}
	return nil
}

// CreateClient implements [usecase.AuthUserRepo].
func (u *UserRepo) CreateClient(ctx context.Context, client domain.CreateClient, tx pgx.Tx) error {
	const query = `INSERT INTO "clients"("user_uuid", "info") VALUES ($1, $2);`
	_, err := u.repo.GetDB(tx).Exec(ctx, query, client.UserUUID, client.Info)
	if err != nil {
		return err
	}
	return nil
}
