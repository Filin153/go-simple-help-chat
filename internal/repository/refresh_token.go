package repository

import (
	"context"
	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type RefreshTokenRepo struct {
	repo *Repository
}

func NewRefreshTokenRepo(repo *Repository) *RefreshTokenRepo {
	return &RefreshTokenRepo{
		repo: repo,
	}
}

func (r *RefreshTokenRepo) Get(ctx context.Context, jti string, tx pgx.Tx) (*domain.RefreshToken, error) {
	const query = `SELECT * FROM "refresh_tokens" WHERE jti = $1;`
	rows, err := r.repo.GetDB(tx).Query(ctx, query, jti)
	if err != nil {
		return &domain.RefreshToken{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.RefreshToken])
	if err != nil {
		return &domain.RefreshToken{}, err
	}

	return &res, nil
}

func (r *RefreshTokenRepo) Create(ctx context.Context, jti, userUUID string, tx pgx.Tx) error {
	const query = `INSERT INTO "refresh_tokens"("jti", "user_uuid") VALUES ($1, $2);`
	_, err := r.repo.GetDB(tx).Exec(ctx, query, jti, userUUID)
	if err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokenRepo) DeleteByUserUUID(ctx context.Context, userUUID string, tx pgx.Tx) error {
	const query = `DELETE FROM "refresh_tokens" WHERE user_uuid = $1;`
	_, err := r.repo.GetDB(tx).Exec(ctx, query, userUUID)
	if err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokenRepo) DeleteByJTI(ctx context.Context, jti string, tx pgx.Tx) error {
	const query = `DELETE FROM "refresh_tokens" WHERE jti = $1;`
	_, err := r.repo.GetDB(tx).Exec(ctx, query, jti)
	if err != nil {
		return err
	}
	return nil
}
