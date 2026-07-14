package domain

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleClient  UserRole = "client"
)

type CreateUser struct {
	Login    string   `json:"login" validate:"required"`
	Password string   `json:"password" validate:"required,min=6"`
	Role     UserRole `json:"-"`
}

type UpdateUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     UserRole
}

type User struct {
	UUID         string   `db:"uuid" json:"uuid"`
	Login        string   `db:"login" json:"login"`
	Password     string   `db:"password" json:"-"`
	Role         UserRole `db:"role" json:"role"`
	Name         *string  `db:"name" json:"name,omitempty"`
	DepartmentID *int     `db:"department_id" json:"department_id,omitempty"`
}

type UserFilter struct {
	Role   UserRole
	Search string
}

type Manager struct {
	ID           int    `db:"id" json:"id"`
	DepartmentID int    `db:"department_id" json:"department_id"`
	UserUUID     string `db:"user_uuid" json:"user_uuid"`
	Name         string `db:"name" json:"name"`
}

type CreateManager struct {
	CreateUser
	DepartmentID int    `json:"department_id" validate:"required"`
	UserUUID     string `json:"user_uuid,omitempty"`
	Name         string `json:"name" validate:"required"`
}

type Client struct {
	ID       int            `db:"id" json:"id"`
	UserUUID string         `db:"user_uuid" json:"user_uuid"`
	Info     map[string]any `db:"info" json:"info"`
}

type CreateClient struct {
	CreateUser
	UserUUID string         `db:"user_uuid" json:"user_uuid"`
	Info     map[string]any `db:"info" json:"info"`
}

type UserSystemInfo struct {
	UUID     string
	UserRole UserRole
	Scope    []string
}
