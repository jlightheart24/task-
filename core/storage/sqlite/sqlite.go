package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"taskpp/core/crypto"
	"taskpp/core/model"

	_ "modernc.org/sqlite"
)

// Store is a SQLite-backed storage implementation.
type Store struct {
	dsn string
	db  *sql.DB
	enc crypto.Cryptor
}

// New creates a new Store for the provided DSN.
func New(dsn string, enc crypto.Cryptor) *Store {
	return &Store{dsn: dsn, enc: enc}
}

// Open opens the database and runs migrations.
func (s *Store) Open() error {
	if s.db != nil {
		return nil
	}
	conn, err := sql.Open("sqlite", s.dsn)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	s.db = conn
	if err := s.migrate(context.Background()); err != nil {
		_ = conn.Close()
		s.db = nil
		return err
	}
	return nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

// Conn exposes the underlying connection (for tests).
func (s *Store) Conn() *sql.DB {
	return s.db
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			ciphertext BLOB NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS task_events (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			seq INTEGER NOT NULL,
			ts TEXT NOT NULL,
			type TEXT NOT NULL,
			payload BLOB NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sync_state (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			last_seq INTEGER NOT NULL,
			last_sync TEXT NOT NULL,
			device_id TEXT NOT NULL,
			server_tag TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS key_state (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			salt BLOB NOT NULL,
			kdf TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func (s *Store) ListTasks(filter model.TaskFilter) ([]model.Task, error) {
	if err := s.Open(); err != nil {
		return nil, err
	}
	if s.enc == nil || !s.enc.IsUnlocked() {
		return nil, fmt.Errorf("keys not unlocked")
	}
	if err := s.ensureEncryptedTasks(context.Background()); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`SELECT id, ciphertext FROM tasks`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	out := make([]model.Task, 0)
	for rows.Next() {
		var id string
		var ciphertext []byte
		if err := rows.Scan(&id, &ciphertext); err != nil {
			return nil, fmt.Errorf("list tasks scan: %w", err)
		}
		payload, err := s.enc.Decrypt(ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decrypt task: %w", err)
		}
		task, err := decodeTask(payload)
		if err != nil {
			return nil, fmt.Errorf("decode task: %w", err)
		}
		if task.ID == "" {
			task.ID = id
		}
		if !matchesFilter(task, filter) {
			continue
		}
		out = append(out, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tasks rows: %w", err)
	}
	sortTasks(out)
	return out, nil
}

func (s *Store) GetTask(id string) (model.Task, error) {
	if err := s.Open(); err != nil {
		return model.Task{}, err
	}
	if s.enc == nil || !s.enc.IsUnlocked() {
		return model.Task{}, fmt.Errorf("keys not unlocked")
	}
	if err := s.ensureEncryptedTasks(context.Background()); err != nil {
		return model.Task{}, err
	}

	row := s.db.QueryRow(`SELECT id, ciphertext FROM tasks WHERE id = ?`, id)
	var task model.Task
	var ciphertext []byte
	var storedID string
	if err := row.Scan(&storedID, &ciphertext); err != nil {
		if err == sql.ErrNoRows {
			return model.Task{}, nil
		}
		return model.Task{}, fmt.Errorf("get task: %w", err)
	}
	payload, err := s.enc.Decrypt(ciphertext)
	if err != nil {
		return model.Task{}, fmt.Errorf("decrypt task: %w", err)
	}
	task, err = decodeTask(payload)
	if err != nil {
		return model.Task{}, fmt.Errorf("decode task: %w", err)
	}
	if task.ID == "" {
		task.ID = storedID
	}
	return task, nil
}

func (s *Store) UpsertTask(task model.Task) error {
	if err := s.Open(); err != nil {
		return err
	}
	if s.enc == nil || !s.enc.IsUnlocked() {
		return fmt.Errorf("keys not unlocked")
	}
	if err := s.ensureEncryptedTasks(context.Background()); err != nil {
		return err
	}

	stmt := `INSERT INTO tasks (
		id, ciphertext
	) VALUES (?, ?)
	ON CONFLICT(id) DO UPDATE SET
		ciphertext = excluded.ciphertext`

	payload, err := encodeTask(task)
	if err != nil {
		return fmt.Errorf("encode task: %w", err)
	}
	ciphertext, err := s.enc.Encrypt(payload)
	if err != nil {
		return fmt.Errorf("encrypt task: %w", err)
	}
	_, err = s.db.Exec(
		stmt,
		task.ID,
		ciphertext,
	)
	if err != nil {
		return fmt.Errorf("upsert task: %w", err)
	}
	return nil
}

func (s *Store) DeleteTask(id string) error {
	if err := s.Open(); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (s *Store) AppendEvents(events []model.Event) error {
	if err := s.Open(); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("append events begin: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO task_events (id, device_id, seq, ts, type, payload) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("append events prepare: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		if _, err := stmt.Exec(
			event.ID,
			event.DeviceID,
			event.Seq,
			formatTime(event.TS),
			event.Type,
			event.Payload,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("append events exec: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("append events commit: %w", err)
	}
	return nil
}

func (s *Store) ListEventsSince(seq int64) ([]model.Event, error) {
	if err := s.Open(); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`SELECT id, device_id, seq, ts, type, payload FROM task_events WHERE seq > ? ORDER BY seq ASC`, seq)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	out := make([]model.Event, 0)
	for rows.Next() {
		var event model.Event
		var ts string
		if err := rows.Scan(&event.ID, &event.DeviceID, &event.Seq, &ts, &event.Type, &event.Payload); err != nil {
			return nil, fmt.Errorf("list events scan: %w", err)
		}
		var err error
		event.TS, err = parseTime(ts)
		if err != nil {
			return nil, fmt.Errorf("parse event ts: %w", err)
		}
		out = append(out, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list events rows: %w", err)
	}
	return out, nil
}

func (s *Store) GetKeyState() (model.KeyState, error) {
	if err := s.Open(); err != nil {
		return model.KeyState{}, err
	}
	row := s.db.QueryRow(`SELECT salt, kdf, updated_at FROM key_state WHERE id = 1`)
	var state model.KeyState
	var updatedAt string
	if err := row.Scan(&state.Salt, &state.KDF, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return model.KeyState{}, nil
		}
		return model.KeyState{}, fmt.Errorf("get key state: %w", err)
	}
	parsed, err := parseTime(updatedAt)
	if err != nil {
		return model.KeyState{}, fmt.Errorf("parse key state time: %w", err)
	}
	state.UpdatedAt = parsed
	return state, nil
}

func (s *Store) SaveKeyState(state model.KeyState) error {
	if err := s.Open(); err != nil {
		return err
	}
	stmt := `INSERT INTO key_state (id, salt, kdf, updated_at)
	VALUES (1, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		salt = excluded.salt,
		kdf = excluded.kdf,
		updated_at = excluded.updated_at`
	if _, err := s.db.Exec(
		stmt,
		state.Salt,
		state.KDF,
		formatTime(state.UpdatedAt),
	); err != nil {
		return fmt.Errorf("save key state: %w", err)
	}
	return nil
}

func (s *Store) GetSyncState() (model.SyncState, error) {
	if err := s.Open(); err != nil {
		return model.SyncState{}, err
	}

	row := s.db.QueryRow(`SELECT last_seq, last_sync, device_id, server_tag FROM sync_state WHERE id = 1`)
	var state model.SyncState
	var lastSync string
	if err := row.Scan(&state.LastSeq, &lastSync, &state.DeviceID, &state.ServerTag); err != nil {
		if err == sql.ErrNoRows {
			return model.SyncState{}, nil
		}
		return model.SyncState{}, fmt.Errorf("get sync state: %w", err)
	}
	parsed, err := parseTime(lastSync)
	if err != nil {
		return model.SyncState{}, fmt.Errorf("parse sync state time: %w", err)
	}
	state.LastSync = parsed
	return state, nil
}

func (s *Store) SaveSyncState(state model.SyncState) error {
	if err := s.Open(); err != nil {
		return err
	}

	stmt := `INSERT INTO sync_state (id, last_seq, last_sync, device_id, server_tag)
	VALUES (1, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		last_seq = excluded.last_seq,
		last_sync = excluded.last_sync,
		device_id = excluded.device_id,
		server_tag = excluded.server_tag`

	if _, err := s.db.Exec(
		stmt,
		state.LastSeq,
		formatTime(state.LastSync),
		state.DeviceID,
		state.ServerTag,
	); err != nil {
		return fmt.Errorf("save sync state: %w", err)
	}
	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

func parseTime(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, input)
}

func parseDate(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", input)
}

func encodeTask(task model.Task) ([]byte, error) {
	return json.Marshal(task)
}

func decodeTask(payload []byte) (model.Task, error) {
	var task model.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func matchesFilter(task model.Task, filter model.TaskFilter) bool {
	if filter.Status != "" && task.Status != filter.Status {
		return false
	}
	if filter.Archived != nil && task.Archived != *filter.Archived {
		return false
	}
	if filter.DueDate != "" {
		if formatDate(task.DueDate) != filter.DueDate {
			return false
		}
	}
	return true
}

func sortTasks(tasks []model.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		ai := tasks[i]
		aj := tasks[j]
		aiZero := ai.DueDate.IsZero()
		ajZero := aj.DueDate.IsZero()
		if aiZero != ajZero {
			return !aiZero
		}
		if !aiZero && !ajZero {
			if !ai.DueDate.Equal(aj.DueDate) {
				return ai.DueDate.Before(aj.DueDate)
			}
		}
		if ai.Order != aj.Order {
			return ai.Order < aj.Order
		}
		return ai.CreatedAt.Before(aj.CreatedAt)
	})
}

func (s *Store) ensureEncryptedTasks(ctx context.Context) error {
	hasCiphertext, err := s.hasColumn(ctx, "tasks", "ciphertext")
	if err != nil {
		return err
	}
	if hasCiphertext {
		return nil
	}
	if s.enc == nil || !s.enc.IsUnlocked() {
		return fmt.Errorf("keys not unlocked")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate tasks begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`CREATE TABLE tasks_new (id TEXT PRIMARY KEY, ciphertext BLOB NOT NULL);`); err != nil {
		return fmt.Errorf("migrate tasks create: %w", err)
	}

	rows, err := tx.Query(`SELECT id, title, description, status, priority, due_date, order_index, created_at, updated_at, completed_at, archived FROM tasks`)
	if err != nil {
		return fmt.Errorf("migrate tasks select: %w", err)
	}
	defer rows.Close()

	insertStmt, err := tx.Prepare(`INSERT INTO tasks_new (id, ciphertext) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("migrate tasks prepare: %w", err)
	}
	defer insertStmt.Close()

	for rows.Next() {
		var task model.Task
		var dueDate, createdAt, updatedAt, completedAt string
		var archivedInt int
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&dueDate,
			&task.Order,
			&createdAt,
			&updatedAt,
			&completedAt,
			&archivedInt,
		); err != nil {
			return fmt.Errorf("migrate tasks scan: %w", err)
		}
		var err error
		task.DueDate, err = parseDate(dueDate)
		if err != nil {
			return fmt.Errorf("migrate due_date: %w", err)
		}
		task.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return fmt.Errorf("migrate created_at: %w", err)
		}
		task.UpdatedAt, err = parseTime(updatedAt)
		if err != nil {
			return fmt.Errorf("migrate updated_at: %w", err)
		}
		task.CompletedAt, err = parseTime(completedAt)
		if err != nil {
			return fmt.Errorf("migrate completed_at: %w", err)
		}
		task.Archived = archivedInt == 1

		payload, err := encodeTask(task)
		if err != nil {
			return fmt.Errorf("migrate encode: %w", err)
		}
		ciphertext, err := s.enc.Encrypt(payload)
		if err != nil {
			return fmt.Errorf("migrate encrypt: %w", err)
		}
		if _, err := insertStmt.Exec(task.ID, ciphertext); err != nil {
			return fmt.Errorf("migrate insert: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate rows: %w", err)
	}

	if _, err := tx.Exec(`DROP TABLE tasks`); err != nil {
		return fmt.Errorf("migrate drop: %w", err)
	}
	if _, err := tx.Exec(`ALTER TABLE tasks_new RENAME TO tasks`); err != nil {
		return fmt.Errorf("migrate rename: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migrate commit: %w", err)
	}
	return nil
}

func (s *Store) hasColumn(ctx context.Context, tableName, columnName string) (bool, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, fmt.Errorf("pragma table_info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, colType string
		var notnull int
		var dfltValue any
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return false, fmt.Errorf("pragma scan: %w", err)
		}
		if name == columnName {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("pragma rows: %w", err)
	}
	return false, nil
}
