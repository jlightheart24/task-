package model

import "time"

// Task is the internal core model. Not bind-safe.
type Task struct {
	ID          string
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     time.Time
	Order       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt time.Time
	Archived    bool
}
