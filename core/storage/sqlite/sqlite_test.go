package sqlite

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"taskpp/core/crypto"
	"taskpp/core/model"

	_ "modernc.org/sqlite"
)

func TestOpenAndMigrate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core.db")
	store := New("file:"+path, newTestCryptor(t))
	if err := store.Open(); err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	tables := []string{"tasks", "task_events", "sync_state", "key_state"}
	for _, table := range tables {
		var name string
		err := store.Conn().QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
}

func TestUpsertListDeleteTasks(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	now := time.Now().UTC().Truncate(time.Second)
	task := model.Task{
		ID:          "t1",
		Title:       "First",
		Description: "desc",
		Status:      "active",
		Priority:    "med",
		DueDate:     now,
		Order:       1,
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: time.Time{},
		Archived:    false,
	}

	if err := store.UpsertTask(task); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	rows, err := store.ListTasks(model.TaskFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 task, got %d", len(rows))
	}

	task.Title = "Updated"
	task.UpdatedAt = now.Add(time.Minute)
	if err := store.UpsertTask(task); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	rows, err = store.ListTasks(model.TaskFilter{})
	if err != nil {
		t.Fatalf("list after update: %v", err)
	}
	if rows[0].Title != "Updated" {
		t.Fatalf("expected updated title, got %q", rows[0].Title)
	}

	if err := store.DeleteTask(task.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	rows, err = store.ListTasks(model.TaskFilter{})
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 tasks after delete, got %d", len(rows))
	}
}

func TestEventsAndSyncState(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	events := []model.Event{
		{ID: "e1", DeviceID: "d1", Seq: 1, TS: time.Now().UTC(), Type: "create", Payload: []byte("one")},
		{ID: "e2", DeviceID: "d1", Seq: 2, TS: time.Now().UTC(), Type: "update", Payload: []byte("two")},
	}
	if err := store.AppendEvents(events); err != nil {
		t.Fatalf("append events: %v", err)
	}

	rows, err := store.ListEventsSince(0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 events, got %d", len(rows))
	}

	state := model.SyncState{
		LastSeq:   2,
		LastSync:  time.Now().UTC(),
		DeviceID:  "d1",
		ServerTag: "s1",
	}
	if err := store.SaveSyncState(state); err != nil {
		t.Fatalf("save sync state: %v", err)
	}
	loaded, err := store.GetSyncState()
	if err != nil {
		t.Fatalf("get sync state: %v", err)
	}
	if loaded.LastSeq != 2 || loaded.DeviceID != "d1" || loaded.ServerTag != "s1" {
		t.Fatalf("unexpected sync state: %+v", loaded)
	}
}

func TestMigrationFromPlaintextTasks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core.db")
	conn, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		status TEXT NOT NULL,
		priority TEXT NOT NULL,
		due_date TEXT NOT NULL,
		order_index INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		completed_at TEXT NOT NULL,
		archived INTEGER NOT NULL
	);`)
	if err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	_, err = conn.Exec(
		`INSERT INTO tasks (id, title, description, status, priority, due_date, order_index, created_at, updated_at, completed_at, archived)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"t1", "Old", "desc", "active", "med", now.Format("2006-01-02"), 1,
		now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), "", 0,
	)
	if err != nil {
		t.Fatalf("insert old row: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close conn: %v", err)
	}

	store := New("file:"+path, newTestCryptor(t))
	if err := store.Open(); err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	tasks, err := store.ListTasks(model.TaskFilter{})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Old" {
		t.Fatalf("unexpected title: %q", tasks[0].Title)
	}
}

func TestConflictsRoundTrip(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	now := time.Now().UTC().Truncate(time.Second)
	conflict := model.Conflict{
		ID:             "c1",
		TaskID:         "t1",
		LocalUpdatedAt: now,
		RemoteEventID:  "e1",
		RemoteTS:       now.Add(-time.Minute),
		DetectedAt:     now,
		Resolution:     "lww_local",
	}
	if err := store.AddConflict(conflict); err != nil {
		t.Fatalf("add conflict: %v", err)
	}
	conflicts, err := store.ListConflicts()
	if err != nil {
		t.Fatalf("list conflicts: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].TaskID != "t1" {
		t.Fatalf("unexpected conflict task id: %s", conflicts[0].TaskID)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "core.db")
	store := New("file:"+path, newTestCryptor(t))
	if err := store.Open(); err != nil {
		t.Fatalf("open: %v", err)
	}
	if store.Conn() == nil {
		t.Fatalf("expected db connection")
	}
	if err := store.Conn().QueryRow("SELECT 1").Scan(new(int)); err != nil {
		t.Fatalf("db ping: %v", err)
	}
	return store
}

func newTestCryptor(t *testing.T) *crypto.Manager {
	t.Helper()
	manager := crypto.NewManager()
	salt := []byte("0123456789abcdef")
	if err := manager.DeriveKey("passphrase", salt); err != nil {
		t.Fatalf("derive key: %v", err)
	}
	return manager
}
