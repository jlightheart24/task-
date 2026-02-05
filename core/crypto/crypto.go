package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
)

// Cryptor is the minimal interface used by storage.
type Cryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	IsUnlocked() bool
}

// Manager manages an in-memory encryption key.
type Manager struct {
	key []byte
}

// NewManager creates an empty manager.
func NewManager() *Manager {
	return &Manager{}
}

// IsUnlocked returns true when a key is loaded.
func (m *Manager) IsUnlocked() bool {
	return len(m.key) == chacha20poly1305.KeySize
}

// DeriveKey derives and sets the key from a passphrase and salt.
func (m *Manager) DeriveKey(passphrase string, salt []byte) error {
	if len(salt) == 0 {
		return fmt.Errorf("salt is required")
	}
	key, err := scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, chacha20poly1305.KeySize)
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}
	m.key = key
	return nil
}

// Encrypt encrypts plaintext with XChaCha20-Poly1305.
func (m *Manager) Encrypt(plaintext []byte) ([]byte, error) {
	if !m.IsUnlocked() {
		return nil, fmt.Errorf("keys not unlocked")
	}
	aead, err := chacha20poly1305.NewX(m.key)
	if err != nil {
		return nil, fmt.Errorf("create aead: %w", err)
	}
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

// Decrypt decrypts ciphertext with XChaCha20-Poly1305.
func (m *Manager) Decrypt(ciphertext []byte) ([]byte, error) {
	if !m.IsUnlocked() {
		return nil, fmt.Errorf("keys not unlocked")
	}
	if len(ciphertext) < chacha20poly1305.NonceSizeX {
		return nil, fmt.Errorf("ciphertext too short")
	}
	aead, err := chacha20poly1305.NewX(m.key)
	if err != nil {
		return nil, fmt.Errorf("create aead: %w", err)
	}
	nonce := ciphertext[:chacha20poly1305.NonceSizeX]
	payload := ciphertext[chacha20poly1305.NonceSizeX:]
	plaintext, err := aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

// NewSalt returns a random 16-byte salt.
func NewSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("salt: %w", err)
	}
	return salt, nil
}
