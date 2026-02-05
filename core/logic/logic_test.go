package logic

import (
	"testing"
	"time"
)

func TestValidateTask(t *testing.T) {
	if err := ValidateTask("", "active", "med", "", zeroTime()); err == nil {
		t.Fatalf("expected error for empty title")
	}
	if err := ValidateTask("Title", "bad", "med", "", zeroTime()); err == nil {
		t.Fatalf("expected error for invalid status")
	}
	if err := ValidateTask("Title", "active", "bad", "", zeroTime()); err == nil {
		t.Fatalf("expected error for invalid priority")
	}
	if err := ValidateTask("Title", "active", "med", "2026-99-99", zeroTime()); err == nil {
		t.Fatalf("expected error for invalid due date")
	}
}

func TestValidateTaskCompletion(t *testing.T) {
	if err := ValidateTask("Title", "done", "med", "", zeroTime()); err == nil {
		t.Fatalf("expected error when done without completed_at")
	}
	if err := ValidateTask("Title", "active", "med", "", nonZeroTime()); err == nil {
		t.Fatalf("expected error when active with completed_at")
	}
}

func zeroTime() time.Time { return time.Time{} }

func nonZeroTime() time.Time {
	return time.Now().UTC()
}
