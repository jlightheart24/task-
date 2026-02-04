package sqlite

import (
	"context"
	"testing"
)

func TestMigrateCreatesTables(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tables := []string{"tasks", "task_events", "conflicts", "devices", "settings"}
	for _, table := range tables {
		var name string
		err := db.Conn().QueryRowContext(
			context.Background(),
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
	}
}
