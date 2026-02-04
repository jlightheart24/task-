package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// DB wraps a SQLite connection.
type DB struct {
	conn *sql.DB
}

// Open opens a SQLite database at the given DSN.
func Open(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	return &DB{conn: conn}, nil
}

// Close closes the underlying connection.
func (db *DB) Close() error {
	if db == nil || db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

// Migrate creates the minimal schema for local storage.
func (db *DB) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			ciphertext BLOB NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			deleted_at INTEGER,
			version INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS task_events (
			event_id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			op TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			device_id TEXT NOT NULL,
			ciphertext BLOB NOT NULL,
			synced INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS conflicts (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			local_event_id TEXT NOT NULL,
			remote_event_id TEXT NOT NULL,
			detected_at INTEGER NOT NULL,
			resolved_at INTEGER,
			resolution TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.conn.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Conn exposes the underlying sql.DB.
func (db *DB) Conn() *sql.DB {
	return db.conn
}
