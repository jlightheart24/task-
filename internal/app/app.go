package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"taskpp/internal/domain"
	"taskpp/internal/storage/sqlite"

	"github.com/google/uuid"
)

type App struct {
	env string
	db  *sqlite.DB
}

// New creates a new app instance for binding into Wails.
func New(env, dsn string) (*App, error) {
	db, err := sqlite.Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &App{env: env, db: db}, nil
}

// Env exposes the current environment for diagnostics.
func (a *App) Env() string {
	return a.env
}

// Greet provides a minimal backend method for UI binding checks.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "there"
	}
	return fmt.Sprintf("hello %s from %s", name, a.env)
}

// CreateTask adds a new task to the local database.
func (a *App) CreateTask(title string, dueDate string) (domain.Task, error) {
	now := time.Now().UTC()
	parsedDueDate, err := parseDueDate(dueDate)
	if err != nil {
		return domain.Task{}, err
	}
	if parsedDueDate.IsZero() {
		parsedDueDate = todayDueDate()
	}
	task := domain.Task{
		ID:        uuid.NewString(),
		Title:     title,
		Description: "",
		Status:    "open",
		DueDate:   parsedDueDate,
		Order:     now.UnixNano(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal task: %w", err)
	}

	_, err = a.db.Conn().ExecContext(
		context.Background(),
		`INSERT INTO tasks (id, ciphertext, created_at, updated_at, deleted_at, version)
		 VALUES (?, ?, ?, ?, NULL, ?)`,
		task.ID,
		payload,
		task.CreatedAt.Unix(),
		task.UpdatedAt.Unix(),
		1,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("insert task: %w", err)
	}

	return task, nil
}

// ListTasks returns all tasks from local storage.
func (a *App) ListTasks() ([]domain.Task, error) {
	rows, err := a.db.Conn().QueryContext(
		context.Background(),
		`SELECT ciphertext FROM tasks ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		var task domain.Task
		if err := json.Unmarshal(payload, &task); err != nil {
			return nil, fmt.Errorf("unmarshal task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, nil
}

// ToggleTaskComplete flips the task status and updates timestamps.
func (a *App) ToggleTaskComplete(id string) (domain.Task, error) {
	var payload []byte
	err := a.db.Conn().QueryRowContext(
		context.Background(),
		`SELECT ciphertext FROM tasks WHERE id = ?`,
		id,
	).Scan(&payload)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return domain.Task{}, fmt.Errorf("unmarshal task: %w", err)
	}

	now := time.Now().UTC()
	if task.Status == "done" {
		task.Status = "open"
		task.CompletedAt = time.Time{}
	} else {
		task.Status = "done"
		task.CompletedAt = now
	}
	task.UpdatedAt = now

	updatedPayload, err := json.Marshal(task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal task: %w", err)
	}

	_, err = a.db.Conn().ExecContext(
		context.Background(),
		`UPDATE tasks SET ciphertext = ?, updated_at = ? WHERE id = ?`,
		updatedPayload,
		task.UpdatedAt.Unix(),
		task.ID,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

// DeleteTask removes a task from local storage.
func (a *App) DeleteTask(id string) error {
	_, err := a.db.Conn().ExecContext(
		context.Background(),
		`DELETE FROM tasks WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// UpdateTaskOrder sets the task's order value for manual reordering.
func (a *App) UpdateTaskOrder(id string, order int64) (domain.Task, error) {
	var payload []byte
	err := a.db.Conn().QueryRowContext(
		context.Background(),
		`SELECT ciphertext FROM tasks WHERE id = ?`,
		id,
	).Scan(&payload)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return domain.Task{}, fmt.Errorf("unmarshal task: %w", err)
	}

	task.Order = order
	task.UpdatedAt = time.Now().UTC()

	updatedPayload, err := json.Marshal(task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal task: %w", err)
	}

	_, err = a.db.Conn().ExecContext(
		context.Background(),
		`UPDATE tasks SET ciphertext = ?, updated_at = ? WHERE id = ?`,
		updatedPayload,
		task.UpdatedAt.Unix(),
		task.ID,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

// UpdateTaskDetails updates description, due date, and priority.
func (a *App) UpdateTaskDetails(id string, description string, dueDate string, priority string) (domain.Task, error) {
	var payload []byte
	err := a.db.Conn().QueryRowContext(
		context.Background(),
		`SELECT ciphertext FROM tasks WHERE id = ?`,
		id,
	).Scan(&payload)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return domain.Task{}, fmt.Errorf("unmarshal task: %w", err)
	}

	parsedDueDate, err := parseDueDate(dueDate)
	if err != nil {
		return domain.Task{}, err
	}

	task.Description = description
	task.DueDate = parsedDueDate
	task.Priority = priority
	task.UpdatedAt = time.Now().UTC()

	updatedPayload, err := json.Marshal(task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal task: %w", err)
	}

	_, err = a.db.Conn().ExecContext(
		context.Background(),
		`UPDATE tasks SET ciphertext = ?, updated_at = ? WHERE id = ?`,
		updatedPayload,
		task.UpdatedAt.Unix(),
		task.ID,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

// UpdateTaskDueDate sets or clears a task's due date.
func (a *App) UpdateTaskDueDate(id string, dueDate string) (domain.Task, error) {
	var payload []byte
	err := a.db.Conn().QueryRowContext(
		context.Background(),
		`SELECT ciphertext FROM tasks WHERE id = ?`,
		id,
	).Scan(&payload)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return domain.Task{}, fmt.Errorf("unmarshal task: %w", err)
	}

	parsedDueDate, err := parseDueDate(dueDate)
	if err != nil {
		return domain.Task{}, err
	}

	task.DueDate = parsedDueDate
	task.UpdatedAt = time.Now().UTC()

	updatedPayload, err := json.Marshal(task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal task: %w", err)
	}

	_, err = a.db.Conn().ExecContext(
		context.Background(),
		`UPDATE tasks SET ciphertext = ?, updated_at = ? WHERE id = ?`,
		updatedPayload,
		task.UpdatedAt.Unix(),
		task.ID,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

func parseDueDate(dueDate string) (time.Time, error) {
	if dueDate == "" {
		return time.Time{}, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", dueDate, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date: %w", err)
	}
	return parsed.UTC(), nil
}

func todayDueDate() time.Time {
	now := time.Now()
	year, month, day := now.Date()
	localMidnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	return localMidnight.UTC()
}
