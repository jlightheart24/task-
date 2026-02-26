package domain

import "time"

// Task is the core unit of work tracked by the app.
// All fields are intended to be encrypted in storage.
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ShortTitle  string `json:"short_title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	Priority    string `json:"priority,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Order       int64  `json:"order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at,omitempty"`
	Archived    bool   `json:"archived"`
}

// Age returns how long the task has existed since creation.
func (t Task) Age(now time.Time) time.Duration {
	createdAt, err := time.Parse(time.RFC3339Nano, t.CreatedAt)
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339, t.CreatedAt)
	}
	if err != nil {
		return 0
	}
	if now.Before(createdAt) {
		return 0
	}
	return now.Sub(createdAt)
}
