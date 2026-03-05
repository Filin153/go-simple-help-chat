package repository

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(ctx context.Context, dns string, logLevel slog.Level) (*Repository, error) {
	config, err := pgxpool.ParseConfig(dns)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	_, err = pool.Exec(ctx, "SELECT 1;")
	if err != nil {
		return nil, err
	}

	return &Repository{
		DB: pool,
	}, nil
}

func (r *Repository) CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return r.DB.BeginTx(ctx, options)
}

func (r *Repository) GetDB(tx pgx.Tx) Execer {
	if tx == nil {
		return r.DB
	}
	return tx
}
