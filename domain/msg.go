package domain

import "time"

type MsgFromType string

const (
	MsgFromTypeManager MsgFromType = "manager"
	MsgFromTypeClient  MsgFromType = "client"
	MsgFromTypeSystem  MsgFromType = "system"
)

type CreateMsgFromClient struct {
	Text  string `json:"text" validate:"required,max=500"`
	Files [][]byte
}

type CreateMsgFromManager struct {
	TicketID int    `json:"ticket_id" validate:"required"`
	Text     string `json:"text" validate:"required,max=500"`
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
	Read     bool             `db:"read" json:"read"`
	CreateAt time.Time        `db:"create_at" json:"create_at"`
	Files    []MsgFileContent `json:"files"`
	Ticket   Ticket           `json:"ticket"`
	FromUser User             `json:"from_user"`
	ToUser   User             `json:"to_user"`
}
