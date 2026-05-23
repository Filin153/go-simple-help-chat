package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type poolPinger interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

var (
	parsePoolConfig   = pgxpool.ParseConfig
	newPoolWithConfig = pgxpool.NewWithConfig
	pingPool          = defaultPingPool
)

type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	DB      *pgxpool.Pool
	beginTx func(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
	db      Execer
}

func defaultPingPool(ctx context.Context, pool poolPinger) error {
	_, err := pool.Exec(ctx, "SELECT 1;")
	return err
}

func NewRepository(ctx context.Context, dns string) (*Repository, error) {
	config, err := parsePoolConfig(dns)
	if err != nil {
		return nil, err
	}

	pool, err := newPoolWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pingPool(ctx, pool); err != nil {
		return nil, err
	}

	return &Repository{
		DB:      pool,
		beginTx: pool.BeginTx,
		db:      pool,
	}, nil
}

func (r *Repository) CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return r.beginTx(ctx, options)
}

func (r *Repository) GetDB(tx pgx.Tx) Execer {
	if tx == nil {
		return r.db
	}
	return tx
}

func getOffset(page, limit int) int {
	return (page - 1) * limit
}
