package repository

import (
	"context"
	"fmt"
	"shc/domain"
	"sort"
	"strings"

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

func (r *Repository) Close() {
	r.DB.Close()
}

func getOffset(page, limit int) int {
	return (page - 1) * limit
}

func getUpdateQuery(table string, updateColumn map[string]any, whereAnd map[string]any) (string, []any, error) {
	if len(updateColumn) == 0 {
		return "", nil, domain.ErrEmptyObject
	}

	table = pgx.Identifier{table}.Sanitize()
	paramNumber := 1
	args := make([]any, 0, len(updateColumn))
	setQueryValues := make([]string, 0, len(updateColumn))
	whereAndQueryValues := make([]string, 0, len(whereAnd))
	query := make([]string, 0, 7)
	query = append(query, "UPDATE")
	query = append(query, table)
	query = append(query, "SET")

	updateColumns := make([]string, 0, len(updateColumn))
	for column := range updateColumn {
		updateColumns = append(updateColumns, column)
	}
	sort.Strings(updateColumns)

	for _, column := range updateColumns {
		arg := updateColumn[column]
		column = pgx.Identifier{column}.Sanitize()
		setQueryValues = append(setQueryValues, fmt.Sprintf("%s = $%d", column, paramNumber))
		args = append(args, arg)
		paramNumber++
	}

	whereColumns := make([]string, 0, len(whereAnd))
	for column := range whereAnd {
		whereColumns = append(whereColumns, column)
	}
	sort.Strings(whereColumns)

	for _, column := range whereColumns {
		arg := whereAnd[column]
		column = pgx.Identifier{column}.Sanitize()
		whereAndQueryValues = append(whereAndQueryValues, fmt.Sprintf("%s = $%d", column, paramNumber))
		args = append(args, arg)
		paramNumber++
	}

	query = append(query, strings.Join(setQueryValues, ", "))
	if len(whereAndQueryValues) > 0 {
		query = append(query, "WHERE")
		query = append(query, strings.Join(whereAndQueryValues, " AND "))
	}

	return strings.Join(query, " ") + ";", args, nil
}
