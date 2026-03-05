package domain

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleClient  UserRole = "client"
)

type CreateUser struct {
	UUID     string
	Login    string
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     UserRole
}

type User struct {
	UUID     string   `db:"uuid" json:"uuid"`
	Login    string   `db:"login" json:"login"`
	Name     string   `db:"name" json:"name"`
	Role     UserRole `db:"role" json:"role"`
	Password string   `db:"password"`
}
