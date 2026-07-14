package domain

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
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Default  bool      `json:"default_dep"`
	Schedule *Schedule `json:"schedule,omitempty"`
}
