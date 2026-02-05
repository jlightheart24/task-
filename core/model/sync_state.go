package model

import "time"

// SyncState tracks the last known sync position.
type SyncState struct {
	LastSeq   int64
	LastSync  time.Time
	DeviceID  string
	ServerTag string
}
