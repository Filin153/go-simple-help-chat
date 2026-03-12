package domain

type RefreshToken struct {
	JTI    string `db:"jti"`
	UserID int    `db:"user_id"`
}
