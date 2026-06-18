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

func (m *MsgRepo) CreateForClient(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
	const query = `INSERT INTO "messages" ("ticket_id", "recipient_user_uuid", "from_type", "text", "status")
SELECT $1, t."client_user_uuid", $2, $3, $4
FROM "tickets" AS t
WHERE t."id" = $1
RETURNING "id", "from_type", "text", "ticket_id", "status", "create_at";`

	return m.create(ctx, query, msg, domain.MsgFromTypeManager, tx)
}

func (m *MsgRepo) CreateForManager(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
	const query = `INSERT INTO "messages" ("ticket_id", "recipient_user_uuid", "from_type", "text", "status")
SELECT $1, t."manager_user_uuid", $2, $3, $4
FROM "tickets" AS t
WHERE t."id" = $1
RETURNING "id", "from_type", "text", "ticket_id", "status", "create_at";`

	return m.create(ctx, query, msg, domain.MsgFromTypeClient, tx)
}

func (m *MsgRepo) CreateFile(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (domain.MsgFileContent, error) {
	const query = `INSERT INTO "message_files" ("msg_id", "file_name", "path")
VALUES ($1, $2, $3)
RETURNING "id", "msg_id", "file_name", "path";`

	rows, err := m.repo.GetDB(tx).Query(ctx, query, msgID, fileName, path)
	if err != nil {
		return domain.MsgFileContent{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.MsgFileContent])
	if err != nil {
		return domain.MsgFileContent{}, err
	}

	return res, nil
}

func (m *MsgRepo) GetUnread(ctx context.Context, userUUID string, tx pgx.Tx) ([]*domain.Msg, error) {
	const query = `SELECT "id", "from_type", "text", "ticket_id", "status", "create_at"
FROM "messages"
WHERE "recipient_user_uuid" = $1 AND "status" = $2
ORDER BY "create_at" ASC, "id" ASC;`

	return m.getMsgs(ctx, tx, query, userUUID, domain.MsgStatusSent)
}

func (m *MsgRepo) MarkReadByID(ctx context.Context, userUUID string, id int, tx pgx.Tx) error {
	const query = `UPDATE "messages"
SET "status" = $3
WHERE "recipient_user_uuid" = $1 AND "id" = $2;`

	tag, err := m.repo.GetDB(tx).Exec(ctx, query, userUUID, id, domain.MsgStatusRead)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUnknownObject
	}

	return nil
}

func (m *MsgRepo) GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]*domain.Msg, error) {
	const query = `SELECT m."id", m."from_type", m."text", m."ticket_id", m."status", m."create_at"
FROM "messages" AS m
JOIN "tickets" AS t ON t."id" = m."ticket_id"
WHERE m."ticket_id" = $2
  AND m."create_at" >= $3
  AND m."create_at" <= $4
  AND (t."client_user_uuid" = $1 OR t."manager_user_uuid" = $1)
ORDER BY m."create_at" ASC, m."id" ASC;`

	return m.getMsgs(ctx, nil, query, userUUID, ticketUUID, from, to)
}

func (m *MsgRepo) create(ctx context.Context, query string, msg domain.CreateMsg, fromType domain.MsgFromType, tx pgx.Tx) (*domain.Msg, error) {
	rows, err := m.repo.GetDB(tx).Query(ctx, query, msg.TicketID, fromType, msg.EncryptedText, domain.MsgStatusSent)
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

func (m *MsgRepo) getMsgs(ctx context.Context, tx pgx.Tx, query string, args ...any) ([]*domain.Msg, error) {
	rows, err := m.repo.GetDB(tx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Msg])
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return []*domain.Msg{}, nil
	}

	filesByMsgID, err := m.getFilesByMsgIDs(ctx, tx, msgIDs(msgs))
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		msg.Files = filesByMsgID[msg.ID]
		if msg.Files == nil {
			msg.Files = []domain.MsgFileContent{}
		}
	}

	return msgs, nil
}

func (m *MsgRepo) getFilesByMsgIDs(ctx context.Context, tx pgx.Tx, ids []int) (map[int][]domain.MsgFileContent, error) {
	if len(ids) == 0 {
		return map[int][]domain.MsgFileContent{}, nil
	}

	const query = `SELECT "id", "msg_id", "file_name", "path"
FROM "message_files"
WHERE "msg_id" = ANY($1)
ORDER BY "msg_id" ASC, "id" ASC;`

	rows, err := m.repo.GetDB(tx).Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.MsgFileContent])
	if err != nil {
		return nil, err
	}

	res := make(map[int][]domain.MsgFileContent, len(ids))
	for _, file := range files {
		res[file.MsgID] = append(res[file.MsgID], file)
	}

	return res, nil
}

func msgIDs(msgs []*domain.Msg) []int {
	res := make([]int, 0, len(msgs))
	for _, msg := range msgs {
		res = append(res, msg.ID)
	}
	return res
}
