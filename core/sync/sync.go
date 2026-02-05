package sync

import (
	"encoding/json"
	"fmt"
	"time"

	"taskpp/core/model"
)

// ApplyEvent applies a single event to the task state using LWW rules.
// Returns (updatedTask, changed, conflict, error).
func ApplyEvent(task model.Task, event model.Event) (model.Task, bool, bool, error) {
	var payload TaskDTO
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return model.Task{}, false, false, fmt.Errorf("decode event payload: %w", err)
	}

	evtTime, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
	if err != nil && payload.UpdatedAt != "" {
		return model.Task{}, false, false, fmt.Errorf("parse updated_at: %w", err)
	}
	if payload.UpdatedAt == "" {
		evtTime = event.TS
	}

	existing := task
	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil && payload.CreatedAt != "" {
		return model.Task{}, false, false, fmt.Errorf("parse created_at: %w", err)
	}
	completedAt, err := time.Parse(time.RFC3339Nano, payload.CompletedAt)
	if err != nil && payload.CompletedAt != "" {
		return model.Task{}, false, false, fmt.Errorf("parse completed_at: %w", err)
	}
	dueDate, err := time.Parse("2006-01-02", payload.DueDate)
	if err != nil && payload.DueDate != "" {
		return model.Task{}, false, false, fmt.Errorf("parse due_date: %w", err)
	}

	updated := model.Task{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Status:      payload.Status,
		Priority:    payload.Priority,
		DueDate:     dueDate,
		Order:       payload.Order,
		CreatedAt:   createdAt,
		UpdatedAt:   evtTime,
		CompletedAt: completedAt,
		Archived:    payload.Archived,
	}

	if existing.ID == "" {
		return updated, true, false, nil
	}
	if updated.UpdatedAt.After(existing.UpdatedAt) {
		return updated, true, false, nil
	}
	if updated.UpdatedAt.Equal(existing.UpdatedAt) && event.Seq > 0 {
		return updated, true, false, nil
	}
	if updated.UpdatedAt.Before(existing.UpdatedAt) {
		return existing, false, true, nil
	}
	return existing, false, false, nil
}

// TaskDTO mirrors bind.TaskDTO without imports to avoid dependency cycles.
type TaskDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
	Order       int64  `json:"order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at"`
	Archived    bool   `json:"archived"`
}
