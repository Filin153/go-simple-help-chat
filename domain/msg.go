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
	TicketID      int    `json:"ticket_id" validate:"required"`
	Text          string `json:"text" validate:"required,max=2500"`
	EncryptedText []byte
	Files         []CreateFile
}

type CreateFile struct {
	Name string `json:"file_name"`
	Data []byte `json:"data"`
	Path string
}

type MsgFileContent struct {
	ID       int    `db:"id" json:"id"`
	MsgID    int    `db:"msg_id" json:"msg_id"`
	FileName string `db:"file_name" json:"file_name"`
	Path     string `db:"path" json:"path"`
}

type Msg struct {
	ID            int              `db:"id" json:"id"`
	FromType      MsgFromType      `db:"from_type" json:"from_type"`
	Text          string           `json:"text"`
	EncryptedText []byte           `db:"text" json:"-"`
	TicketID      int              `db:"ticket_id" json:"ticket_id"`
	Status        MsgStatus        `db:"status" json:"status"`
	CreateAt      time.Time        `db:"create_at" json:"create_at"`
	Files         []MsgFileContent `json:"files"`
}
