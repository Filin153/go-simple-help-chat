package domain

import "time"

type TicketStatus string

const (
	NewTicketStatus            TicketStatus = "new"
	InWorkTicketStatus         TicketStatus = "in_work"
	HaveUnreadTicketStatus     TicketStatus = "have_unread"
	ReadyTicketStatus          TicketStatus = "ready"
	NextDepartmentTicketStatus TicketStatus = "next_department"
)

type CreateTicket struct {
	ClientUserID int
}

type Ticket struct {
	ID            int          `db:"id" json:"id"`
	DepartmentID  int          `db:"department_id" json:"department_id"`
	Name          string       `db:"name" json:"name"`
	ManagerUserID int          `db:"manager_user_id" json:"manager_user_id"`
	ClientUserID  int          `db:"client_user_id" json:"client_user_id"`
	Status        TicketStatus `db:"status" json:"status"`
	CreateAt      time.Time    `db:"create_at" json:"create_at"`
	UpdateAt      time.Time    `db:"update_at" json:"update_at"`
}
