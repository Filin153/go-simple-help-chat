package domain

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleClient  UserRole = "client"
)

type CreateUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     UserRole
}

type UpdateUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     UserRole
}

type User struct {
	UUID     string   `db:"uuid" json:"uuid"`
	Login    string   `db:"login" json:"login"`
	Password string   `db:"password"`
	Role     UserRole `db:"role" json:"role"`
}

type Manager struct {
	ID           int    `db:"id" json:"id"`
	DepartmentID int    `db:"department_id" json:"department_id"`
	UserUUID     string `db:"user_uuid" json:"user_uuid"`
	Name         string `db:"name" json:"name"`
}

type CreateManager struct {
	CreateUser
	DepartmentID int    `json:"department_id"`
	UserUUID     string `json:"user_uuid"`
	Name         string `json:"name"`
}

type Client struct {
	ID       int         `db:"id" json:"id"`
	UserUUID string      `db:"user_uuid" json:"user_uuid"`
	Info     map[string]any `db:"info" json:"info"`
}

type CreateClient struct {
	CreateUser
	UserUUID string      `db:"user_uuid" json:"user_uuid"`
	Info     map[string]any `db:"info" json:"info"`
}

type UserSystemInfo struct {
	UUID     string
	UserRole UserRole
	Scope    []string
}
