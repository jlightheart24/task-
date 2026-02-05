package model

import "time"

// Event represents a single mutation in the sync log.
type Event struct {
	ID       string
	DeviceID string
	Seq      int64
	TS       time.Time
	Type     string
	Payload  []byte
}
