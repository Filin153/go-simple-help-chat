package service

import (
	"errors"
	"testing"

	"github.com/alexedwards/argon2id"
)

func useDefaultPasswordHooks(t *testing.T) {
	t.Helper()

	oldCreateHash := createHash
	oldComparePasswordAndHash := comparePasswordAndHash

	t.Cleanup(func() {
		createHash = oldCreateHash
		comparePasswordAndHash = oldComparePasswordAndHash
	})
}

func Test_PASSWORD_OK(t *testing.T) {
	password := "pa$$word"
	hash, err := (PasswordCoder{}).CreatePasswordHash(password)
	if err != nil {
		t.Fatalf("CreatePasswordHash returned error: %v", err)
	}

	match := (PasswordCoder{}).VerifyPassword(password, hash)

	if !match {
		t.Fatalf("same password must match; got=%t", match)
	}
}

func Test_PASSWORD_MISMATCH(t *testing.T) {
	hash, err := (PasswordCoder{}).CreatePasswordHash("pa$$word")
	if err != nil {
		t.Fatalf("CreatePasswordHash returned error: %v", err)
	}

	match := (PasswordCoder{}).VerifyPassword("not-the-same-password", hash)
	if match {
		t.Fatal("different password must not match")
	}
}

func Test_PASSWORD_INVALID_HASH(t *testing.T) {
	match := (PasswordCoder{}).VerifyPassword("pa$$word", "invalid-hash")
	if match {
		t.Fatal("invalid hash must return false")
	}
}

func Test_PASSWORD_HASH_ERROR(t *testing.T) {
	useDefaultPasswordHooks(t)
	createHash = func(_ string, _ *argon2id.Params) (string, error) {
		return "", errors.New("create hash error")
	}

	hash, err := (PasswordCoder{}).CreatePasswordHash("pa$$word")
	if err == nil {
		t.Fatal("CreatePasswordHash expected error")
	}
	if hash != "" {
		t.Fatalf("expected empty hash on error, got=%q", hash)
	}
}

func Test_PASSWORD_VERIFY_ERROR(t *testing.T) {
	useDefaultPasswordHooks(t)
	comparePasswordAndHash = func(_, _ string) (bool, error) {
		return false, errors.New("compare error")
	}

	match := (PasswordCoder{}).VerifyPassword("pa$$word", "hash")
	if match {
		t.Fatal("VerifyPassword must return false on compare error")
	}
}
