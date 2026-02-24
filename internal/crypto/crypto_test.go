package crypto

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{0x5a}, 32)
	plaintext := []byte(`{"id":"1","title":"hello"}`)

	enc, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	dec, err := Decrypt(key, enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(dec, plaintext) {
		t.Fatalf("expected round-trip to match plaintext")
	}
}

func TestDecryptUnknownCiphertext(t *testing.T) {
	key := bytes.Repeat([]byte{0x5a}, 32)
	_, err := Decrypt(key, []byte(`{"id":"1"}`))
	if !errors.Is(err, ErrUnknownCiphertext) {
		t.Fatalf("expected ErrUnknownCiphertext, got %v", err)
	}
}
