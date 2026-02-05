package logic

import (
	"fmt"
	"time"
)

// ValidateTask enforces basic domain rules.
func ValidateTask(title, status, priority, dueDate string, completedAt time.Time) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	switch status {
	case "", "active", "done":
	default:
		return fmt.Errorf("invalid status: %s", status)
	}
	switch priority {
	case "", "low", "med", "high":
	default:
		return fmt.Errorf("invalid priority: %s", priority)
	}
	if dueDate != "" {
		if _, err := time.Parse("2006-01-02", dueDate); err != nil {
			return fmt.Errorf("invalid due_date: %s", dueDate)
		}
	}
	if status == "done" && completedAt.IsZero() {
		return fmt.Errorf("completed_at required when status=done")
	}
	if status != "done" && !completedAt.IsZero() {
		return fmt.Errorf("completed_at must be empty when status!=done")
	}
	return nil
}
