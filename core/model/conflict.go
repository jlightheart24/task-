package model

import "time"

// Conflict records a detected sync conflict.
type Conflict struct {
	ID             string
	TaskID         string
	LocalUpdatedAt time.Time
	RemoteEventID  string
	RemoteTS       time.Time
	DetectedAt     time.Time
	Resolution     string
}
