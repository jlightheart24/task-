package bind

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestCoreCreateAndList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bind.db")
	cfg := Config{
		StorageDriver: "sqlite",
		StoragePath:   "file:" + path,
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	core := NewCore(string(cfgJSON))
	if errStr := core.Open(); errStr != "" {
		t.Fatalf("open: %s", errStr)
	}
	defer core.Close()

	if errStr := core.InitKeys("passphrase"); errStr != "" {
		t.Fatalf("init keys: %s", errStr)
	}

	created := core.CreateTask(`{"title":"One","description":"d"}`)
	if hasError(created) {
		t.Fatalf("create error: %s", created)
	}

	events := core.ExportEvents(0)
	if hasError(events) {
		t.Fatalf("export error: %s", events)
	}
	var evs []EventDTO
	if err := json.Unmarshal([]byte(events), &evs); err != nil {
		t.Fatalf("decode events: %v", err)
	}
	if len(evs) != 1 || evs[0].Type != "create" {
		t.Fatalf("expected create event, got %+v", evs)
	}

	listed := core.ListTasks("")
	if hasError(listed) {
		t.Fatalf("list error: %s", listed)
	}

	var tasks []TaskDTO
	if err := json.Unmarshal([]byte(listed), &tasks); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "One" {
		t.Fatalf("unexpected title: %q", tasks[0].Title)
	}
}

func TestCoreValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bind.db")
	cfg := Config{
		StorageDriver: "sqlite",
		StoragePath:   "file:" + path,
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	core := NewCore(string(cfgJSON))
	if errStr := core.Open(); errStr != "" {
		t.Fatalf("open: %s", errStr)
	}
	defer core.Close()

	if errStr := core.InitKeys("passphrase"); errStr != "" {
		t.Fatalf("init keys: %s", errStr)
	}

	created := core.CreateTask(`{"title":""}`)
	if !hasError(created) {
		t.Fatalf("expected validation error, got: %s", created)
	}
}

func hasError(payload string) bool {
	var errObj map[string]string
	if err := json.Unmarshal([]byte(payload), &errObj); err != nil {
		return false
	}
	_, ok := errObj["error"]
	return ok
}
