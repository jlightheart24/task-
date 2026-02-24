package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	appcrypto "taskpp/internal/crypto"
	"taskpp/internal/domain"
	"taskpp/internal/storage/sqlite"

	"github.com/google/uuid"
)

type App struct {
	env string
	db  *sqlite.DB
	dek []byte
}

// New creates a new app instance for binding into Wails.
func New(env, dsn string) (*App, error) {
	db, err := sqlite.Open(dsn)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	dek, err := loadOrCreateDEK(ctx, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	app := &App{env: env, db: db, dek: dek}
	if err := app.migrateTasksToEncrypted(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return app, nil
}

// Env exposes the current environment for diagnostics.
func (a *App) Env() string {
	return a.env
}

// Close releases any underlying resources.
func (a *App) Close() error {
	if a == nil || a.db == nil {
		return nil
	}
	return a.db.Close()
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
		ID:          uuid.NewString(),
		Title:       title,
		Description: "",
		Status:      "open",
		DueDate:     parsedDueDate,
		Order:       now.UnixNano(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal task: %w", err)
	}

	encrypted, err := a.encryptPayload(payload)
	if err != nil {
		return domain.Task{}, err
	}

	_, err = a.db.Conn().ExecContext(
		context.Background(),
		`INSERT INTO tasks (id, ciphertext, created_at, updated_at, deleted_at, version)
		 VALUES (?, ?, ?, ?, NULL, ?)`,
		task.ID,
		encrypted,
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
		`SELECT id, ciphertext FROM tasks ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var (
			id      string
			payload []byte
		)
		if err := rows.Scan(&id, &payload); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		task, err := a.decryptTaskPayload(context.Background(), id, payload)
		if err != nil {
			return nil, err
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
	ctx := context.Background()
	task, err := a.loadTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
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

	if err := a.saveTask(ctx, task); err != nil {
		return domain.Task{}, err
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
	ctx := context.Background()
	task, err := a.loadTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	task.Order = order
	task.UpdatedAt = time.Now().UTC()

	if err := a.saveTask(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

// UpdateTaskDetails updates description, due date, and priority.
func (a *App) UpdateTaskDetails(id string, title string, description string, dueDate string, priority string) (domain.Task, error) {
	ctx := context.Background()
	task, err := a.loadTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	parsedDueDate, err := parseDueDate(dueDate)
	if err != nil {
		return domain.Task{}, err
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle != "" {
		task.Title = trimmedTitle
	}
	task.Description = description
	task.DueDate = parsedDueDate
	task.Priority = priority
	task.UpdatedAt = time.Now().UTC()

	if err := a.saveTask(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

// UpdateTaskDueDate sets or clears a task's due date.
func (a *App) UpdateTaskDueDate(id string, dueDate string) (domain.Task, error) {
	ctx := context.Background()
	task, err := a.loadTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	parsedDueDate, err := parseDueDate(dueDate)
	if err != nil {
		return domain.Task{}, err
	}

	task.DueDate = parsedDueDate
	task.UpdatedAt = time.Now().UTC()

	if err := a.saveTask(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

const dekSettingKey = "dek_v1"

func loadOrCreateDEK(ctx context.Context, db *sqlite.DB) ([]byte, error) {
	value, ok, err := db.GetSetting(ctx, dekSettingKey)
	if err != nil {
		return nil, err
	}
	if ok {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("decode dek: %w", err)
		}
		if len(decoded) != 32 {
			return nil, fmt.Errorf("decode dek: invalid key length %d", len(decoded))
		}
		return decoded, nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate dek: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := db.SetSetting(ctx, dekSettingKey, encoded); err != nil {
		return nil, err
	}
	return key, nil
}

func (a *App) encryptPayload(plaintext []byte) ([]byte, error) {
	encrypted, err := appcrypto.Encrypt(a.dek, plaintext)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}
	return encrypted, nil
}

func (a *App) decryptTaskPayload(ctx context.Context, id string, payload []byte) (domain.Task, error) {
	plaintext, err := appcrypto.Decrypt(a.dek, payload)
	if err == nil {
		var task domain.Task
		if err := json.Unmarshal(plaintext, &task); err != nil {
			return domain.Task{}, fmt.Errorf("unmarshal task: %w", err)
		}
		return task, nil
	}
	if errors.Is(err, appcrypto.ErrUnknownCiphertext) {
		var task domain.Task
		if jsonErr := json.Unmarshal(payload, &task); jsonErr != nil {
			return domain.Task{}, fmt.Errorf("decrypt task: %w", err)
		}
		if task.ID == "" {
			task.ID = id
		}
		if id != "" && task.ID != id {
			return domain.Task{}, fmt.Errorf("migrate task: id mismatch")
		}
		if err := a.persistEncryptedTask(ctx, task); err != nil {
			return domain.Task{}, err
		}
		return task, nil
	}
	return domain.Task{}, fmt.Errorf("decrypt task: %w", err)
}

func (a *App) loadTask(ctx context.Context, id string) (domain.Task, error) {
	var payload []byte
	err := a.db.Conn().QueryRowContext(
		ctx,
		`SELECT ciphertext FROM tasks WHERE id = ?`,
		id,
	).Scan(&payload)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}
	return a.decryptTaskPayload(ctx, id, payload)
}

func (a *App) saveTask(ctx context.Context, task domain.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}
	encrypted, err := a.encryptPayload(payload)
	if err != nil {
		return err
	}
	_, err = a.db.Conn().ExecContext(
		ctx,
		`UPDATE tasks SET ciphertext = ?, updated_at = ? WHERE id = ?`,
		encrypted,
		task.UpdatedAt.Unix(),
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (a *App) persistEncryptedTask(ctx context.Context, task domain.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}
	encrypted, err := a.encryptPayload(payload)
	if err != nil {
		return err
	}
	_, err = a.db.Conn().ExecContext(
		ctx,
		`UPDATE tasks SET ciphertext = ? WHERE id = ?`,
		encrypted,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (a *App) migrateTasksToEncrypted(ctx context.Context) error {
	rows, err := a.db.Conn().QueryContext(
		ctx,
		`SELECT id, ciphertext FROM tasks`,
	)
	if err != nil {
		return fmt.Errorf("migrate tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id      string
			payload []byte
		)
		if err := rows.Scan(&id, &payload); err != nil {
			return fmt.Errorf("migrate tasks: scan: %w", err)
		}
		if _, err := appcrypto.Decrypt(a.dek, payload); err == nil {
			continue
		} else if errors.Is(err, appcrypto.ErrUnknownCiphertext) {
			var task domain.Task
			if jsonErr := json.Unmarshal(payload, &task); jsonErr != nil {
				return fmt.Errorf("migrate tasks: decrypt: %w", err)
			}
			if task.ID == "" {
				task.ID = id
			}
			if task.ID != id {
				return fmt.Errorf("migrate tasks: id mismatch")
			}
			if err := a.persistEncryptedTask(ctx, task); err != nil {
				return err
			}
			continue
		} else {
			return fmt.Errorf("migrate tasks: decrypt: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate tasks: %w", err)
	}
	return nil
}

func parseDueDate(dueDate string) (time.Time, error) {
	if dueDate == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, dueDate); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, dueDate); err == nil {
		return parsed, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", dueDate, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date: %w", err)
	}
	return parsed, nil
}

func todayDueDate() time.Time {
	now := time.Now()
	year, month, day := now.Date()
	localMidnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	return localMidnight
}
