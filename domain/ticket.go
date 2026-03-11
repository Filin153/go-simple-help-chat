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
	ClientUserUUID string
}

type Ticket struct {
	ID              int          `db:"id" json:"id"`
	DepartmentID    int          `db:"department_id" json:"department_id"`
	Name            string       `db:"name" json:"name"`
	ManagerUserUUID string       `db:"manager_user_uuid" json:"manager_user_uuid"`
	ClientUserUUID  string       `db:"client_user_uuid" json:"client_user_uuid"`
	Status          TicketStatus `db:"status" json:"status"`
	CreateAt        time.Time    `db:"create_at" json:"create_at"`
	UpdateAt        time.Time    `db:"update_at" json:"update_at"`
}
