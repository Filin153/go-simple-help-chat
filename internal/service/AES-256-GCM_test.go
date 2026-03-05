package service

import (
	"bytes"
	"crypto/cipher"
	"errors"
	"io"
	"strings"
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

func Test_AES256GCM_OK(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	msg := []byte("AES-256-GCM")
	aad := []byte("aad")

	enc, err := EncryptAES256GCM(key, msg, aad)
	if err != nil {
		t.Fatalf("EncryptAES256GCM returned error: %v", err)
	}

	dec, err := DecryptAES256GCM(key, enc, aad)
	if err != nil {
		t.Fatalf("DecryptAES256GCM returned error: %v", err)
	}

	if !bytes.Equal(msg, dec) {
		t.Fatalf("plaintext mismatch; got %q want %q", dec, msg)
	}
}

func Test_AES256GCM_DiffAdd_ERROR(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	msg := []byte("AES-256-GCM")

	enc, err := EncryptAES256GCM(key, msg, []byte("aad"))
	if err != nil {
		t.Fatalf("EncryptAES256GCM returned error: %v", err)
	}

	_, err = DecryptAES256GCM(key, enc, []byte(""))
	if err == nil {
		t.Fatal("DecryptAES256GCM expected error for mismatched aad")
	}
}

func Test_AES256GCM_DiffKey_ERROR(t *testing.T) {
	keyOne := bytes.Repeat([]byte{1}, 32)
	keyTwo := bytes.Repeat([]byte{2}, 32)
	msg := []byte("AES-256-GCM")
	aad := []byte("aad")

	enc, err := EncryptAES256GCM(keyOne, msg, aad)
	if err != nil {
		t.Fatalf("EncryptAES256GCM returned error: %v", err)
	}

	_, err = DecryptAES256GCM(keyTwo, enc, aad)
	if err == nil {
		t.Fatal("DecryptAES256GCM expected error for mismatched key")
	}
}

func Test_AES256GCM_Encrypt_InvalidKeyLength(t *testing.T) {
	_, err := EncryptAES256GCM([]byte("short"), []byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("EncryptAES256GCM expected invalid key length error")
	}
	if !strings.Contains(err.Error(), "key must be 32 bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Decrypt_InvalidKeyLength(t *testing.T) {
	_, err := DecryptAES256GCM([]byte("short"), []byte("ciphertext"), []byte("aad"))
	if err == nil {
		t.Fatal("DecryptAES256GCM expected invalid key length error")
	}
	if !strings.Contains(err.Error(), "key must be 32 bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Decrypt_CiphertextTooShort(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)

	_, err := DecryptAES256GCM(key, []byte("short"), []byte("aad"))
	if err == nil {
		t.Fatal("DecryptAES256GCM expected ciphertext too short error")
	}
	if err.Error() != "ciphertext too short" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Encrypt_ReadFullError(t *testing.T) {
	useDefaultCryptoHooks(t)

	readFull = func(_ io.Reader, _ []byte) (int, error) {
		return 0, errors.New("read error")
	}

	_, err := EncryptAES256GCM(bytes.Repeat([]byte{1}, 32), []byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("EncryptAES256GCM expected readFull error")
	}
	if err.Error() != "read error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Encrypt_AESNewCipherError(t *testing.T) {
	useDefaultCryptoHooks(t)

	aesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errors.New("new cipher error")
	}

	_, err := EncryptAES256GCM(bytes.Repeat([]byte{1}, 32), []byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("EncryptAES256GCM expected aesNewCipher error")
	}
	if err.Error() != "new cipher error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Decrypt_AESNewCipherError(t *testing.T) {
	useDefaultCryptoHooks(t)

	aesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errors.New("new cipher error")
	}

	_, err := DecryptAES256GCM(bytes.Repeat([]byte{1}, 32), []byte("ciphertext"), []byte("aad"))
	if err == nil {
		t.Fatal("DecryptAES256GCM expected aesNewCipher error")
	}
	if err.Error() != "new cipher error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Encrypt_CipherNewGCMError(t *testing.T) {
	useDefaultCryptoHooks(t)

	cipherNewGCM = func(_ cipher.Block) (cipher.AEAD, error) {
		return nil, errors.New("new gcm error")
	}

	_, err := EncryptAES256GCM(bytes.Repeat([]byte{1}, 32), []byte("msg"), []byte("aad"))
	if err == nil {
		t.Fatal("EncryptAES256GCM expected cipherNewGCM error")
	}
	if err.Error() != "new gcm error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_AES256GCM_Decrypt_CipherNewGCMError(t *testing.T) {
	useDefaultCryptoHooks(t)

	cipherNewGCM = func(_ cipher.Block) (cipher.AEAD, error) {
		return nil, errors.New("new gcm error")
	}

	_, err := DecryptAES256GCM(bytes.Repeat([]byte{1}, 32), []byte("ciphertext"), []byte("aad"))
	if err == nil {
		t.Fatal("DecryptAES256GCM expected cipherNewGCM error")
	}
	if err.Error() != "new gcm error" {
		t.Fatalf("unexpected error: %v", err)
	}
}
