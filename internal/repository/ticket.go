package repository

import (
	"context"
	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type TicketRepo struct {
	repo *Repository
}

func NewTicketRepo(repo *Repository) *TicketRepo {
	return &TicketRepo{
		repo: repo,
	}
}

func (t *TicketRepo) GetByID(ctx context.Context, id int) (domain.Ticket, error) {
	const query = `SELECT
	"id",
	COALESCE("department_id", 0) AS "department_id",
	"name",
	COALESCE("manager_user_uuid"::TEXT, '') AS "manager_user_uuid",
	"client_user_uuid"::TEXT AS "client_user_uuid",
	"status",
	"create_at",
	"update_at"
FROM "tickets"
WHERE "id" = $1;`

	rows, err := t.repo.GetDB(nil).Query(ctx, query, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Ticket])
	if err != nil {
		return domain.Ticket{}, err
	}

	return res, nil
}
