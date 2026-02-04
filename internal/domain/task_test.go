package domain

import (
	"testing"
	"time"
)

func TestTaskAge(t *testing.T) {
	created := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 2, 4, 10, 0, 0, 0, time.UTC)

	task := Task{CreatedAt: created}
	age := task.Age(now)

	if age != 72*time.Hour {
		t.Fatalf("expected 72h age, got %v", age)
	}
}
