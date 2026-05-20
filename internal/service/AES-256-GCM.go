package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var (
	aesNewCipher = aes.NewCipher
	cipherNewGCM = cipher.NewGCM
	readFull     = io.ReadFull
)

type AES256GCM struct {
	key32 []byte
}

func NewAES256GCM(key32 [32]byte) *AES256GCM {
	return &AES256GCM{
		key32: key32[:],
	}
}

// Encrypt
func (a *AES256GCM) Encrypt(plaintext []byte, aad []byte) ([]byte, error) {
	block, err := aesNewCipher(a.key32)
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

// Decrypt
func (a *AES256GCM) Decrypt(data []byte, aad []byte) ([]byte, error) {
	block, err := aesNewCipher(a.key32)
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
