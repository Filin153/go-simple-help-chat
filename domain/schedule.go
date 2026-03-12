package domain

import (
	"strings"
)

type ScheduleType string

const (
	ScheduleTypeDay   ScheduleType = "day"
	ScheduleTypeBreak ScheduleType = "break"
)

type CreateScheduleBreak struct {
	Name  string `json:"name" validate:"required,max=10"`
	From  string `json:"from" validate:"required"`
	To    string `json:"to" validate:"required"`
	DayID int    `json:"day_id,omitempty" validate:"required"`
}

func (c *CreateScheduleBreak) ValidateTime() error {
	if !strings.Contains(c.From, ":") || !strings.Contains(c.To, ":") {
		return ErrInvalidTimeFormat
	}

	fromSplit := strings.Split(c.From, ":")
	toSplit := strings.Split(c.To, ":")

	if len(fromSplit) != 2 {
		return ErrInvalidTimeFormat
	} else if len(fromSplit[0]) != 2 || len(fromSplit[1]) != 2 {
		return ErrInvalidTimeFormat
	} else if len(toSplit) != 2 {
		return ErrInvalidTimeFormat
	} else if len(toSplit[0]) != 2 || len(toSplit[1]) != 2 {
		return ErrInvalidTimeFormat
	}

	return nil
}

type Schedule struct {
	ID           int          `db:"id" json:"id"`
	DepartmentID int          `db:"department_id" json:"department_id"`
	Name         string       `db:"name" json:"name"`
	From         string       `db:"from" json:"from"`
	To           string       `db:"to" json:"to"`
	Type         ScheduleType `db:"type" json:"type"`
	DayID        int          `db:"day_id" json:"day_id,omitempty"`
	IsDayOff     bool         `db:"is_day_off" json:"is_day_off"`
	Breaks       []Schedule   `db:"-" json:"breaks,omitempty"`
}
