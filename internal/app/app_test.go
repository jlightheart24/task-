package app

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	if app.Env() != "dev" {
		t.Fatalf("expected env to be dev, got %q", app.Env())
	}
}

func TestGreet(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	msg := app.Greet("jonny")
	if msg != "hello jonny from dev" {
		t.Fatalf("unexpected greet message: %q", msg)
	}
}

func TestCreateAndListTasks(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	task, err := app.CreateTask("first", "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.ID == "" {
		t.Fatalf("expected task id to be set")
	}
	if task.Status != "open" {
		t.Fatalf("expected status open, got %q", task.Status)
	}

	tasks, err := app.ListTasks()
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "first" {
		t.Fatalf("unexpected task title: %q", tasks[0].Title)
	}
}

func TestToggleTaskComplete(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	task, err := app.CreateTask("toggle me", "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	updated, err := app.ToggleTaskComplete(task.ID)
	if err != nil {
		t.Fatalf("toggle task: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("expected status done, got %q", updated.Status)
	}
	if updated.CompletedAt == "" {
		t.Fatalf("expected completed_at set")
	}

	updated, err = app.ToggleTaskComplete(task.ID)
	if err != nil {
		t.Fatalf("toggle task: %v", err)
	}
	if updated.Status != "open" {
		t.Fatalf("expected status open, got %q", updated.Status)
	}
	if updated.CompletedAt != "" {
		t.Fatalf("expected completed_at cleared")
	}
}

func TestDeleteTask(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	task, err := app.CreateTask("delete me", "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := app.DeleteTask(task.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	tasks, err := app.ListTasks()
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestUpdateTaskOrder(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	task, err := app.CreateTask("order me", "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	updated, err := app.UpdateTaskOrder(task.ID, 42)
	if err != nil {
		t.Fatalf("update task order: %v", err)
	}
	if updated.Order != 42 {
		t.Fatalf("expected order 42, got %d", updated.Order)
	}
}

func TestUpdateTaskDetails(t *testing.T) {
	dbPath := tempDB(t)
	app, err := New("dev", dbPath)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	task, err := app.CreateTask("details", "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	updated, err := app.UpdateTaskDetails(task.ID, "details", "short", "notes", "2026-02-10", "high")
	if err != nil {
		t.Fatalf("update task details: %v", err)
	}
	if updated.Description != "notes" {
		t.Fatalf("expected description updated, got %q", updated.Description)
	}
	if updated.Priority != "high" {
		t.Fatalf("expected priority high, got %q", updated.Priority)
	}
	if updated.DueDate == "" {
		t.Fatalf("expected due date set")
	}
}

func tempDB(t *testing.T) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "taskminus-*.db")
	if err != nil {
		t.Fatalf("temp db: %v", err)
	}
	_ = file.Close()
	return "file:" + file.Name() + "?_pragma=foreign_keys(1)"
}
