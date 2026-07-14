package repository

import (
	"context"
	"encoding/json"
	"shc/domain"
	"shc/internal/service"

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

func (u *UserRepo) GetAll(ctx context.Context, page, limit int, filter domain.UserFilter, tx pgx.Tx) ([]domain.User, error) {
	const query = `SELECT u.uuid,u.login,u.password,u.role,m.name,m.department_id
		FROM users u LEFT JOIN managers m ON m.user_uuid=u.uuid
		WHERE u.role <> 'client'
		AND ($3 = '' OR u.role::text = $3)
		AND ($4 = '' OR (u.role = 'manager' AND m.name ILIKE '%' || $4 || '%'))
		ORDER BY u.uuid ASC LIMIT $1 OFFSET $2`
	rows, err := u.repo.GetDB(tx).Query(ctx, query, limit, getOffset(page, limit), filter.Role, filter.Search)
	if err != nil {
		return []domain.User{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (u *UserRepo) GetByUUID(ctx context.Context, uuid string, tx pgx.Tx) (*domain.User, error) {
	const query = `SELECT u.uuid,u.login,u.password,u.role,m.name,m.department_id
		FROM users u LEFT JOIN managers m ON m.user_uuid=u.uuid WHERE u.uuid=$1`
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
	const query = `SELECT u.uuid,u.login,u.password,u.role,m.name,m.department_id
		FROM users u LEFT JOIN managers m ON m.user_uuid=u.uuid WHERE u.login=$1`
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

func (u *UserRepo) CreateAdmin(ctx context.Context, user domain.CreateUser, tx pgx.Tx) error {
	const query = `INSERT INTO "users"("login", "password", "role") VALUES ($1, $2, $3);`
	_, err := u.repo.GetDB(tx).Exec(ctx, query, user.Login, user.Password, user.Role)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserRepo) CreateManager(ctx context.Context, manager domain.CreateManager, tx pgx.Tx) error {
	if tx == nil {
		return domain.ErrEmptyObject
	}

	const queryCreateUser = `INSERT INTO "users"("login", "password", "role") VALUES ($1, $2, $3) RETURNING uuid;`
	const queryCreateManager = `INSERT INTO "managers"("department_id", "user_uuid", "name") VALUES ($1, $2, $3);`
	var userUUID string

	err := u.repo.GetDB(tx).QueryRow(ctx, queryCreateUser, manager.Login, manager.Password, manager.Role).Scan(&userUUID)
	if err != nil {
		return err
	}

	_, err = u.repo.GetDB(tx).Exec(ctx, queryCreateManager, manager.DepartmentID, userUUID, manager.Name)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepo) UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser, tx pgx.Tx) error {
	updateCol := service.StructToMap(user, []string{})
	query, args, err := getUpdateQuery("users", updateCol, map[string]any{
		"uuid": uuid,
	})
	if err != nil {
		return err
	}
	tag, err := u.repo.GetDB(tx).Exec(ctx, query, args...)
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
	if tx == nil {
		return domain.ErrEmptyObject
	}

	const queryCreateUser = `INSERT INTO "users"("uuid", "login", "password", "role") VALUES ($1, $2, $3, $4);`
	const queryCreateClient = `INSERT INTO "clients"("user_uuid", "info") VALUES ($1, $2);`

	_, err := u.repo.GetDB(tx).Exec(ctx, queryCreateUser, client.UserUUID, client.Login, client.Password, client.Role)
	if err != nil {
		return err
	}

	userInfoJSON, err := json.Marshal(client.Info)
	if err != nil {
		return err
	}

	_, err = u.repo.GetDB(tx).Exec(ctx, queryCreateClient, client.UserUUID, userInfoJSON)
	if err != nil {
		return err
	}

	return nil
}
