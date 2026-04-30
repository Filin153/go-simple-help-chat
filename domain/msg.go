package domain

import "time"

type MsgFromType string
type MsgStatus string

const (
	MsgFromTypeManager MsgFromType = "manager"
	MsgFromTypeClient  MsgFromType = "client"
	MsgFromTypeSystem  MsgFromType = "system"
)

const (
	MsgStatusDraft MsgStatus = "draft" // сообщение набрано, но ещё не отправлено.
	MsgStatusSent  MsgStatus = "sent"  // сообщение сохраненно на сервер.
	MsgStatusRead  MsgStatus = "read"  // пользователь открыл чат и сообщение отмечено прочитанным.
)

type CreateMsg struct {
	TicketID int    `json:"ticket_id" validate:"required"`
	Text     string `json:"text" validate:"required,max=1000"`
	Files    [][]byte
}

type MsgFileContent struct {
	ID    int    `db:"id" json:"id"`
	MsgID int    `db:"msg_id" json:"msg_id"`
	Path  string `db:"path" json:"path"`
}

type Msg struct {
	ID       int              `db:"id" json:"id"`
	FromType MsgFromType      `db:"from_type" json:"from_type"`
	Text     string           `db:"text" json:"text"`
	TicketID int              `db:"ticket_id" json:"ticket_id"`
	Status   MsgStatus        `db:"status" json:"status"`
	CreateAt time.Time        `db:"create_at" json:"create_at"`
	Files    []MsgFileContent `json:"files"`
}
