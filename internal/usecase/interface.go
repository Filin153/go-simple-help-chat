package usecase

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	oneMonthDuration = time.Hour * 24 * 31
)

// MainRepo creates transactional sessions for use cases.
type MainRepo interface {
	CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}
