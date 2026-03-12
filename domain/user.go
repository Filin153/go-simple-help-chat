package domain

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleClient  UserRole = "client"
)

type CreateUser struct {
	Login    string
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     UserRole
}

type User struct {
	ID       int      `db:"id" json:"id"`
	Login    string   `db:"login" json:"login"`
	Password string   `db:"password"`
	Role     UserRole `db:"role" json:"role"`
}

type Manager struct {
	ID           int    `db:"id" json:"id"`
	DepartmentID int    `db:"department_id" json:"department_id"`
	UserID       int    `db:"user_id" json:"user_id"`
	Name         string `db:"name" json:"name"`
}

type Client struct {
	ID     int         `db:"id" json:"id"`
	UserID int         `db:"user_id" json:"user_id"`
	Info   map[any]any `db:"info" json:"info"`
}

type CreateClient struct {
	UserID int         `db:"user_id" json:"user_id"`
	Info   map[any]any `db:"info" json:"info"`
}
