package service

import (
	"bytes"
	"crypto/cipher"
	"errors"
	"io"
	"testing"
)

func useDefaultCryptoHooks(t *testing.T) {
	t.Helper()

	oldAesNewCipher := aesNewCipher
	oldCipherNewGCM := cipherNewGCM
	oldReadFull := readFull

	t.Cleanup(func() {
		aesNewCipher = oldAesNewCipher
		cipherNewGCM = oldCipherNewGCM
		readFull = oldReadFull
	})
}

func newAESFixture() (*AES256GCM, [32]byte) {
	var key [32]byte
	for i := range len(key) {
		key[i] = 1
	}
	return NewAES256GCM(key), key
}

func Test_NewAES256GCM(t *testing.T) {
	enc, key := newAESFixture()

	if enc == nil {
		t.Fatal("NewAES256GCM returned nil")
	}
	if !bytes.Equal(enc.key32, key[:]) {
		t.Fatalf("unexpected key bytes: got=%v want=%v", enc.key32, key[:])
	}
}

func Test_AES256GCM_OK(t *testing.T) {
	enc, _ := newAESFixture()
	msg := []byte("AES-256-GCM")
	aad := []byte("aad")

	ciphertext, err := enc.Encrypt(msg, aad)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	plaintext, err := enc.Decrypt(ciphertext, aad)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if !bytes.Equal(plaintext, msg) {
		t.Fatalf("plaintext mismatch: got=%q want=%q", plaintext, msg)
	}
}

func Test_AES256GCM_DiffAad_Error(t *testing.T) {
	enc, _ := newAESFixture()
	msg := []byte("AES-256-GCM")

	ciphertext, err := enc.Encrypt(msg, []byte("aad"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	_, err = enc.Decrypt(ciphertext, []byte("other"))
	if err == nil {
		t.Fatal("expected error for mismatched aad")
	}
}

func Test_AES256GCM_DiffKey_Error(t *testing.T) {
	encOne, _ := newAESFixture()
	var keyTwo [32]byte
	for i := range len(keyTwo) {
		keyTwo[i] = 2
	}
	encTwo := NewAES256GCM(keyTwo)
	msg := []byte("AES-256-GCM")
	aad := []byte("aad")

	ciphertext, err := encOne.Encrypt(msg, aad)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	_, err = encTwo.Decrypt(ciphertext, aad)
	if err == nil {
		t.Fatal("expected error for mismatched key")
	}
}

func Test_AES256GCM_Decrypt_CiphertextTooShort(t *testing.T) {
	enc, _ := newAESFixture()

	_, err := enc.Decrypt([]byte("short"), []byte("aad"))
	if err == nil {
		t.Fatal("expected ciphertext too short error")
	}
	if err.Error() != "ciphertext too short" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Encrypt_ReadFullError(t *testing.T) {
	useDefaultCryptoHooks(t)
	enc, _ := newAESFixture()

	readFull = func(_ io.Reader, _ []byte) (int, error) {
		return 0, errors.New("read error")
	}

	_, err := enc.Encrypt([]byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("expected readFull error")
	}
	if err.Error() != "read error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Encrypt_AESNewCipherError(t *testing.T) {
	useDefaultCryptoHooks(t)
	enc, _ := newAESFixture()

	aesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errors.New("new cipher error")
	}

	_, err := enc.Encrypt([]byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("expected aesNewCipher error")
	}
	if err.Error() != "new cipher error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Decrypt_AESNewCipherError(t *testing.T) {
	useDefaultCryptoHooks(t)
	enc, _ := newAESFixture()

	aesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errors.New("new cipher error")
	}

	_, err := enc.Decrypt([]byte("ciphertext"), []byte("aad"))
	if err == nil {
		t.Fatal("expected aesNewCipher error")
	}
	if err.Error() != "new cipher error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Encrypt_CipherNewGCMError(t *testing.T) {
	useDefaultCryptoHooks(t)
	enc, _ := newAESFixture()

	cipherNewGCM = func(_ cipher.Block) (cipher.AEAD, error) {
		return nil, errors.New("new gcm error")
	}

	_, err := enc.Encrypt([]byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("expected cipherNewGCM error")
	}
	if err.Error() != "new gcm error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Decrypt_CipherNewGCMError(t *testing.T) {
	useDefaultCryptoHooks(t)
	enc, _ := newAESFixture()

	cipherNewGCM = func(_ cipher.Block) (cipher.AEAD, error) {
		return nil, errors.New("new gcm error")
	}

	_, err := enc.Decrypt([]byte("ciphertext"), []byte("aad"))
	if err == nil {
		t.Fatal("expected cipherNewGCM error")
	}
	if err.Error() != "new gcm error" {
		t.Fatalf("unexpected error: %v", err)
	}
}
