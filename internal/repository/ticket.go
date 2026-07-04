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

func (t *TicketRepo) Create(ctx context.Context, createTicket domain.CreateTicket) (int, error) {
	const query = `INSERT INTO tickets (
    department_id,
    name,
    client_user_uuid,
    manager_user_uuid
)
VALUES (
    (SELECT id FROM departments WHERE default_dep = true LIMIT 1),
    (SELECT (COUNT(id) + 1)::text FROM tickets WHERE client_user_uuid = $1),
    $1,
    (
        SELECT m.user_uuid
        FROM managers m
        LEFT JOIN tickets t
            ON t.manager_user_uuid = m.user_uuid
           AND t.status NOT IN ('closed', 'resolved')
        WHERE m.department_id = (
            SELECT id FROM departments WHERE default_dep = true LIMIT 1
        )
        GROUP BY m.id, m.user_uuid
        ORDER BY COUNT(t.id) ASC, m.id ASC
        LIMIT 1
    )
) RETURNING id;`
	var id int
	err := t.repo.GetDB(nil).QueryRow(ctx, query, createTicket.ClientUserUUID).Scan(&id)
	return id, err
}

func (t *TicketRepo) GetAll(ctx context.Context, page, limit int) ([]domain.Ticket, error) {
	const query = `SELECT * FROM tickets OFFSET $1 LIMIT $2;`
	rows, err := t.repo.GetDB(nil).Query(ctx, query, getOffset(page, limit), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Ticket])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (t *TicketRepo) GetAllByUserUUID(ctx context.Context, page, limit int, uuid string) ([]domain.Ticket, error) {
	const query = `SELECT * FROM tickets WHERE "client_user_uuid"=$1 OR "manager_user_uuid"=$1 OFFSET $2 LIMIT $3;`
	rows, err := t.repo.GetDB(nil).Query(ctx, query, uuid, getOffset(page, limit), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Ticket])
	if err != nil {
		return nil, err
	}

	return res, nil
}
