package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	keySize   = 32
	versionV1 = byte(1)
)

var (
	ErrUnknownCiphertext = errors.New("unknown ciphertext format")
)

// Encrypt encrypts plaintext with AES-256-GCM and prefixes the payload with a version byte.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("encrypt: invalid key length %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("encrypt: nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, 1+len(nonce)+len(ciphertext))
	out = append(out, versionV1)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

// Decrypt decrypts a versioned AES-256-GCM ciphertext.
func Decrypt(key, data []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("decrypt: invalid key length %d", len(key))
	}
	if len(data) < 2 {
		return nil, ErrUnknownCiphertext
	}
	if data[0] != versionV1 {
		return nil, ErrUnknownCiphertext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < 1+nonceSize {
		return nil, ErrUnknownCiphertext
	}

	nonce := data[1 : 1+nonceSize]
	ciphertext := data[1+nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}
