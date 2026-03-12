package domain

type RefreshToken struct {
	JTI      string `db:"jti"`
	UserUUID string `db:"user_uuid"`
}
