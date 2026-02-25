package domain

import "time"

// Task is the core unit of work tracked by the app.
// All fields are intended to be encrypted in storage.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ShortTitle  string    `json:"short_title,omitempty"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority,omitempty"`
	DueDate     time.Time `json:"due_date,omitempty"`
	Order       int64     `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Archived    bool      `json:"archived"`
}

// Age returns how long the task has existed since creation.
func (t Task) Age(now time.Time) time.Duration {
	if t.CreatedAt.IsZero() {
		return 0
	}
	if now.Before(t.CreatedAt) {
		return 0
	}
	return now.Sub(t.CreatedAt)
}
