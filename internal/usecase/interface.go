package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type MainRepo interface {
	CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}
