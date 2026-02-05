package crypto

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	manager := NewManager()
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("salt: %v", err)
	}
	if err := manager.DeriveKey("passphrase", salt); err != nil {
		t.Fatalf("derive: %v", err)
	}

	plaintext := []byte("hello world")
	ciphertext, err := manager.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	out, err := manager.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(out) != string(plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, out)
	}
}
