package bind

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestImportEventsDedupes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bind.db")
	cfg := Config{StorageDriver: "sqlite", StoragePath: "file:" + path}
	cfgJSON, _ := json.Marshal(cfg)
	core := NewCore(string(cfgJSON))
	if errStr := core.Open(); errStr != "" {
		t.Fatalf("open: %s", errStr)
	}
	defer core.Close()
	if errStr := core.InitKeys("passphrase"); errStr != "" {
		t.Fatalf("init keys: %s", errStr)
	}

	created := core.CreateTask(`{"title":"One"}`)
	if hasError(created) {
		t.Fatalf("create: %s", created)
	}

	exported := core.ExportEvents(0)
	if hasError(exported) {
		t.Fatalf("export: %s", exported)
	}

	var events []EventDTO
	if err := json.Unmarshal([]byte(exported), &events); err != nil {
		t.Fatalf("decode events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	reimportPayload, _ := json.Marshal([]EventDTO{events[0], events[0]})
	if errStr := core.ImportEvents(string(reimportPayload)); errStr != "" {
		t.Fatalf("import: %s", errStr)
	}
}
