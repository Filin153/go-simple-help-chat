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
	ID      int    `db:"id" json:"id"`
	MsgUUID string `db:"msg_uuid" json:"msg_uuid"`
	URL     string `db:"url" json:"url"`
}

type Msg struct {
	UUID     string           `db:"uuid" json:"uuid"`
	TicketID int              `db:"ticket_id" json:"ticket_id"`
	FromType MsgFromType      `db:"from_type" json:"from_type"`
	Text     string           `db:"text" json:"text"`
	CreateAt time.Time        `db:"create_at" json:"create_at"`
	Files    []MsgFileContent `json:"files"`
	Ticket   Ticket           `json:"ticket"`
	FromUser User             `json:"from_user"`
	ToUser   User             `json:"to_user"`
}
