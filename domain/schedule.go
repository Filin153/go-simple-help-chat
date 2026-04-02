package domain

import (
	"time"
)

type CreateScheduleBreak struct {
	Name      string    `json:"name" validate:"required,max=10"`
	From      time.Time `json:"from" validate:"required"`
	To        time.Time `json:"to" validate:"required"`
	IsBreak   bool      `json:"is_break"`
	IsWeekEnd bool      `json:"is_week_end"`
}

func (c *CreateScheduleBreak) Validate() error {
	if c.IsBreak && c.IsWeekEnd {
		return ErrInvalidDayType
	}
	return nil
}

type Schedule struct {
	ID           int       `db:"id" json:"id"`
	DepartmentID int       `db:"department_id" json:"department_id"`
	Name         string    `db:"name" json:"name"`
	From         time.Time `db:"from" json:"from"`
	To           time.Time `db:"to" json:"to"`
	IsBreak      bool      `db:"is_break" json:"is_break"`
	IsWeekEnd    bool      `db:"is_week_end" json:"is_week_end"`
}
