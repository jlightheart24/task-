package model

import "time"

// KeyState stores encryption metadata.
type KeyState struct {
	Salt      []byte
	KDF       string
	UpdatedAt time.Time
}
