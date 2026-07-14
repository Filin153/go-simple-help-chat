package domain

import (
	"time"
)


type CreateSchedule struct {
	DepartmentID int       `json:"department_id" validate:"required"`
	Name         string    `json:"name" validate:"required"`
	WorkFrom     time.Time `json:"work_from" validate:"required"`
	WorkTo       time.Time `json:"work_to" validate:"required"`
	BreakFrom    time.Time `json:"break_from" validate:"required"`
	BreakTo      time.Time `json:"break_to" validate:"required"`
	IsWeekEnd    bool      `json:"is_week_end"`
}

type UpdateSchedule struct {
	ID        int       `json:"id" validate:"required"`
	Name      string    `json:"name"`
	WorkFrom  time.Time `json:"work_from"`
	WorkTo    time.Time `json:"work_to"`
	BreakFrom time.Time `json:"break_from"`
	BreakTo   time.Time `json:"break_to"`
	IsWeekEnd bool      `json:"is_week_end"`
}

type Schedule struct {
	ID           int       `db:"id" json:"id"`
	DepartmentID int       `db:"department_id" json:"department_id"`
	Name         string    `db:"name" json:"name"`
	WorkFrom     time.Time `db:"work_from" json:"work_from"`
	WorkTo       time.Time `db:"work_to" json:"work_to"`
	BreakFrom    time.Time `db:"break_from" json:"break_from"`
	BreakTo      time.Time `db:"break_to" json:"break_to"`
	IsWeekEnd    bool      `db:"is_week_end" json:"is_week_end"`
}
