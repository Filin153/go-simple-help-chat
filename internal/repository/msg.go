package repository

import (
	"context"
	"shc/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type MsgRepo struct {
	repo *Repository
}

func NewMsgRepo(repo *Repository) *MsgRepo {
	return &MsgRepo{
		repo: repo,
	}
}

func (m *MsgRepo) Create(ctx context.Context, msg domain.CreateMsg, userUUID string, tx pgx.Tx) (*domain.Msg, error) {
	const query = `INSERT INTO "messages"("user_uuid", "text", "ticket_id", "status") VALUES ($1, $2, $3, $4) RETURNING *;`
	rows, err := m.repo.GetDB(tx).Query(ctx, query, userUUID, msg.EncryptedText, msg.TicketID, domain.MsgStatusSent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Msg])
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (m *MsgRepo) CreateFile(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (*domain.MsgFileContent, error) {
	const query = `INSERT INTO "message_files"("msg_id", "file_name", "path") VALUES ($1, $2, $3) RETURNING *;`
	rows, err := m.repo.GetDB(tx).Query(ctx, query, msgID, fileName, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.MsgFileContent])
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (m *MsgRepo) GetUnread(ctx context.Context, userUUID string, tx pgx.Tx) ([]domain.Msg, error) {
	const query = `SELECT * FROM "messages" WHERE "user_uuid"=$1 AND "status"=$2;`

	rows, err := m.repo.GetDB(tx).Query(ctx, query, userUUID, domain.MsgStatusSent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Msg])
	return res, nil
}
func (m *MsgRepo) MarkReadByID(ctx context.Context, userUUID string, id int, tx pgx.Tx) error {
	const query = `UPDATE "messages" SET "status"=$1 WHERE "user_uuid"=$2 AND "id"=$3 AND "status"=$4;`
	tag, err := m.repo.GetDB(tx).Exec(ctx, query, domain.MsgStatusRead, userUUID, id, domain.MsgStatusSent)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrZeroRowAffected
	}

	return nil
}
func (m *MsgRepo) GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]domain.Msg, error) {
	const query = `SELECT * FROM "messages" WHERE "user_uuid"=$1 AND "ticket_id"=$2 AND "create_at" >= $3 AND "create_at" < $4;`

	rows, err := m.repo.GetDB(nil).Query(ctx, query, userUUID, ticketUUID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Msg])
	return res, nil
}
