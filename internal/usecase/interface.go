package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// MainRepo creates transactional sessions for use cases.
type MainRepo interface {
	CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}
