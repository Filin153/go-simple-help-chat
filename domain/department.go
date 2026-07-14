package domain

import "time"

type Department struct {
	ID      int    `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	Default bool   `db:"default_dep" json:"default_dep"` // Может быть тольок 1
}

type CreateDepartment struct {
	Name string `json:"name" validate:"required"`
}

type UpdateDepartment struct {
	Name    string `json:"name"`
	Default bool   `json:"default_dep"`
}

type DepartmentWithOnScheduleeDay struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Default   bool      `db:"default_dep" json:"default_dep"`
	WorkFrom  time.Time `db:"work_from" json:"work_from"`
	WorkTo    time.Time `db:"work_to" json:"work_to"`
	IsWeekEnd bool      `db:"is_week_end" json:"is_week_end"`
}
