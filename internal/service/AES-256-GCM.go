package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var (
	aesNewCipher = aes.NewCipher
	cipherNewGCM = cipher.NewGCM
	readFull     = io.ReadFull
)

func EncryptAES256GCM(key32 []byte, plaintext []byte, aad []byte) ([]byte, error) {
	if len(key32) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes (got %d)", len(key32))
	}

	block, err := aesNewCipher(key32)
	if err != nil {
		return nil, err
	}

	gcm, err := cipherNewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := readFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// ciphertext includes auth tag at the end
	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)

	// output: nonce || ciphertext
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func DecryptAES256GCM(key32 []byte, data []byte, aad []byte) ([]byte, error) {
	if len(key32) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes (got %d)", len(key32))
	}

	block, err := aesNewCipher(key32)
	if err != nil {
		return nil, err
	}

	gcm, err := cipherNewGCM(block)
	if err != nil {
		return nil, err
	}

	ns := gcm.NonceSize()
	if len(data) < ns {
		return nil, errors.New("ciphertext too short")
	}

	nonce := data[:ns]
	ciphertext := data[ns:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
