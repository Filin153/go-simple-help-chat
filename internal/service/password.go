package service

import (
	"log/slog"

	"github.com/alexedwards/argon2id"
)

var (
	createHash             = argon2id.CreateHash
	comparePasswordAndHash = argon2id.ComparePasswordAndHash
)

type PasswordCoder struct{}

// CreatePasswordHash builds an Argon2id hash for the password.
func (PasswordCoder) CreatePasswordHash(password string) (string, error) {
	hash, err := createHash(password, argon2id.DefaultParams)
	if err != nil {
		slog.Error("CreateHash", "error", err)
		return "", err
	}
	return hash, nil
}

// VerifyPassword checks that the password matches the stored hash.
func (PasswordCoder) VerifyPassword(password, hashedPassword string) bool {
	match, err := comparePasswordAndHash(password, hashedPassword)
	if err != nil {
		slog.Error("ComparePasswordAndHash", "error", err)
		return false
	}
	return match
}
