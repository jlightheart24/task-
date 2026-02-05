package sync

import (
	"encoding/json"
	"testing"
	"time"

	"taskpp/core/model"
)

func TestApplyEventLWW(t *testing.T) {
	older := model.Task{
		ID:        "t1",
		Title:     "Old",
		UpdatedAt: time.Date(2026, 2, 5, 10, 0, 0, 0, time.UTC),
	}
	event := model.Event{
		ID:  "e1",
		Seq: 1,
		TS:  time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC),
	}
	payload := TaskDTO{
		ID:        "t1",
		Title:     "New",
		UpdatedAt: time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
	}
	data, _ := json.Marshal(payload)
	event.Payload = data

	updated, changed, conflict, err := ApplyEvent(older, event)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if conflict {
		t.Fatalf("did not expect conflict")
	}
	if updated.Title != "New" {
		t.Fatalf("expected title update")
	}
}

func TestApplyEventConflict(t *testing.T) {
	newer := model.Task{
		ID:        "t1",
		Title:     "Newer",
		UpdatedAt: time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC),
	}
	event := model.Event{
		ID:  "e2",
		Seq: 2,
		TS:  time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC),
	}
	payload := TaskDTO{
		ID:        "t1",
		Title:     "Older",
		UpdatedAt: time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
	}
	data, _ := json.Marshal(payload)
	event.Payload = data

	updated, changed, conflict, err := ApplyEvent(newer, event)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if changed {
		t.Fatalf("expected no change")
	}
	if !conflict {
		t.Fatalf("expected conflict")
	}
	if updated.Title != "Newer" {
		t.Fatalf("expected existing task to win")
	}
}
