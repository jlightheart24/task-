package bind

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"taskpp/core/crypto"
	"taskpp/core/model"
	"taskpp/core/logic"
	"taskpp/core/storage"
	"taskpp/core/storage/sqlite"
	"taskpp/core/sync"
)

// Core is the bind-safe facade exposed to UIs.
type Core struct {
	store storage.Storage
	keys  *crypto.Manager
	deviceID string
}

// Config is a bind-safe configuration struct.
type Config struct {
	StorageDriver string `json:"storage_driver"`
	StoragePath   string `json:"storage_path"`
	DeviceID      string `json:"device_id"`
}

// NewCore constructs a core facade from JSON config.
func NewCore(configJSON string) *Core {
	var cfg Config
	if configJSON != "" {
		_ = json.Unmarshal([]byte(configJSON), &cfg)
	}
	keys := crypto.NewManager()
	var store storage.Storage
	switch cfg.StorageDriver {
	case "", "sqlite":
		path := cfg.StoragePath
		if path == "" {
			path = "file:taskpp.db"
		}
		store = sqlite.New(path, keys)
	default:
		store = sqlite.New("file:taskpp.db", keys)
	}
	deviceID := cfg.DeviceID
	if deviceID == "" {
		deviceID = uuid.NewString()
	}
	return &Core{store: store, keys: keys, deviceID: deviceID}
}

// Open initializes the core. Returns empty string on success.
func (c *Core) Open() string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	if err := c.store.Open(); err != nil {
		return errorJSON(fmt.Sprintf("open store: %v", err))
	}
	return ""
}

// Close shuts down the core. Returns empty string on success.
func (c *Core) Close() string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	if err := c.store.Close(); err != nil {
		return errorJSON(fmt.Sprintf("close store: %v", err))
	}
	return ""
}

// TaskDTO is a bind-safe task representation.
type TaskDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
	Order       int64  `json:"order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at"`
	Archived    bool   `json:"archived"`
}

// TaskFilterDTO is a bind-safe filter representation.
type TaskFilterDTO struct {
	Status   string `json:"status"`
	Archived *bool  `json:"archived"`
	DueDate  string `json:"due_date"`
}

// ListTasks returns a JSON-encoded list of TaskDTO.
func (c *Core) ListTasks(filterJSON string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	var filterDTO TaskFilterDTO
	if filterJSON != "" {
		if err := json.Unmarshal([]byte(filterJSON), &filterDTO); err != nil {
			return errorJSON(fmt.Sprintf("decode filter: %v", err))
		}
	}
	tasks, err := c.store.ListTasks(model.TaskFilter{
		Status:   filterDTO.Status,
		Archived: filterDTO.Archived,
		DueDate:  filterDTO.DueDate,
	})
	if err != nil {
		return errorJSON(fmt.Sprintf("list tasks: %v", err))
	}
	out := make([]TaskDTO, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, taskToDTO(task))
	}
	data, err := json.Marshal(out)
	if err != nil {
		return errorJSON(fmt.Sprintf("encode tasks: %v", err))
	}
	return string(data)
}

// CreateTask accepts TaskDTO JSON and returns TaskDTO JSON.
func (c *Core) CreateTask(taskJSON string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	var dto TaskDTO
	if err := json.Unmarshal([]byte(taskJSON), &dto); err != nil {
		return errorJSON(fmt.Sprintf("decode task: %v", err))
	}
	if dto.ID == "" {
		dto.ID = uuid.NewString()
	}
	if dto.Status == "" {
		dto.Status = "active"
	}
	if dto.Priority == "" {
		dto.Priority = "med"
	}
	now := time.Now().UTC()
	if dto.CreatedAt == "" {
		dto.CreatedAt = now.Format(time.RFC3339Nano)
	}
	dto.UpdatedAt = now.Format(time.RFC3339Nano)
	task, err := dtoToTask(dto)
	if err != nil {
		return errorJSON(fmt.Sprintf("convert task: %v", err))
	}
	if err := logic.ValidateTask(dto.Title, dto.Status, dto.Priority, dto.DueDate, task.CompletedAt); err != nil {
		return errorJSON(fmt.Sprintf("validate task: %v", err))
	}
	if err := c.store.UpsertTask(task); err != nil {
		return errorJSON(fmt.Sprintf("create task: %v", err))
	}
	if err := c.appendEvent("create", task); err != nil {
		return errorJSON(fmt.Sprintf("event create: %v", err))
	}
	out, err := json.Marshal(taskToDTO(task))
	if err != nil {
		return errorJSON(fmt.Sprintf("encode task: %v", err))
	}
	return string(out)
}

// UpdateTask accepts TaskDTO JSON and returns TaskDTO JSON.
func (c *Core) UpdateTask(taskJSON string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	var dto TaskDTO
	if err := json.Unmarshal([]byte(taskJSON), &dto); err != nil {
		return errorJSON(fmt.Sprintf("decode task: %v", err))
	}
	if dto.ID == "" {
		return errorJSON("missing id")
	}
	dto.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	task, err := dtoToTask(dto)
	if err != nil {
		return errorJSON(fmt.Sprintf("convert task: %v", err))
	}
	if err := logic.ValidateTask(dto.Title, dto.Status, dto.Priority, dto.DueDate, task.CompletedAt); err != nil {
		return errorJSON(fmt.Sprintf("validate task: %v", err))
	}
	if err := c.store.UpsertTask(task); err != nil {
		return errorJSON(fmt.Sprintf("update task: %v", err))
	}
	if err := c.appendEvent("update", task); err != nil {
		return errorJSON(fmt.Sprintf("event update: %v", err))
	}
	out, err := json.Marshal(taskToDTO(task))
	if err != nil {
		return errorJSON(fmt.Sprintf("encode task: %v", err))
	}
	return string(out)
}

// DeleteTask deletes a task by ID. Returns empty string on success.
func (c *Core) DeleteTask(taskID string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	if taskID == "" {
		return errorJSON("missing id")
	}
	task, err := c.store.GetTask(taskID)
	if err != nil {
		return errorJSON(fmt.Sprintf("load task: %v", err))
	}
	if err := c.store.DeleteTask(taskID); err != nil {
		return errorJSON(fmt.Sprintf("delete task: %v", err))
	}
	if task.ID != "" {
		if err := c.appendEvent("delete", task); err != nil {
			return errorJSON(fmt.Sprintf("event delete: %v", err))
		}
	}
	return ""
}

// ReorderTasks accepts reorder JSON and returns empty string on success.
func (c *Core) ReorderTasks(reorderJSON string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	var items []ReorderItemDTO
	if err := json.Unmarshal([]byte(reorderJSON), &items); err != nil {
		return errorJSON(fmt.Sprintf("decode reorder: %v", err))
	}
	for _, item := range items {
		if item.ID == "" {
			return errorJSON("reorder item missing id")
		}
		task, err := c.store.GetTask(item.ID)
		if err != nil {
			return errorJSON(fmt.Sprintf("load task: %v", err))
		}
		if task.ID == "" {
			return errorJSON(fmt.Sprintf("task not found: %s", item.ID))
		}
		task.Order = item.Order
		if item.DueDate != "" {
			parsed, err := parseDate(item.DueDate)
			if err != nil {
				return errorJSON(fmt.Sprintf("parse due_date: %v", err))
			}
			task.DueDate = parsed
		}
		task.UpdatedAt = time.Now().UTC()
		if err := c.store.UpsertTask(task); err != nil {
			return errorJSON(fmt.Sprintf("reorder task: %v", err))
		}
		if err := c.appendEvent("reorder", task); err != nil {
			return errorJSON(fmt.Sprintf("event reorder: %v", err))
		}
	}
	return ""
}

// SetDueDate updates the due date for a task.
func (c *Core) SetDueDate(taskID string, dueDate string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	if taskID == "" {
		return errorJSON("missing id")
	}
	task, err := c.store.GetTask(taskID)
	if err != nil {
		return errorJSON(fmt.Sprintf("load task: %v", err))
	}
	if task.ID == "" {
		return errorJSON("task not found")
	}
	parsed, err := parseDate(dueDate)
	if err != nil {
		return errorJSON(fmt.Sprintf("parse due_date: %v", err))
	}
	task.DueDate = parsed
	task.UpdatedAt = time.Now().UTC()
	if err := c.store.UpsertTask(task); err != nil {
		return errorJSON(fmt.Sprintf("set due date: %v", err))
	}
	if err := c.appendEvent("set_due_date", task); err != nil {
		return errorJSON(fmt.Sprintf("event set due date: %v", err))
	}
	return ""
}

// SetCompleted updates completion state.
func (c *Core) SetCompleted(taskID string, completed bool) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	if taskID == "" {
		return errorJSON("missing id")
	}
	task, err := c.store.GetTask(taskID)
	if err != nil {
		return errorJSON(fmt.Sprintf("load task: %v", err))
	}
	if task.ID == "" {
		return errorJSON("task not found")
	}
	if completed {
		task.Status = "done"
		task.CompletedAt = time.Now().UTC()
	} else {
		task.Status = "active"
		task.CompletedAt = time.Time{}
	}
	task.UpdatedAt = time.Now().UTC()
	if err := c.store.UpsertTask(task); err != nil {
		return errorJSON(fmt.Sprintf("set completed: %v", err))
	}
	if err := c.appendEvent("set_completed", task); err != nil {
		return errorJSON(fmt.Sprintf("event set completed: %v", err))
	}
	return ""
}

// EventDTO is a bind-safe event representation.
type EventDTO struct {
	ID          string `json:"id"`
	DeviceID    string `json:"device_id"`
	Seq         int64  `json:"seq"`
	TS          string `json:"ts"`
	Type        string `json:"type"`
	PayloadJSON string `json:"payload_json"`
}

// ExportEvents returns JSON-encoded EventDTO list.
func (c *Core) ExportEvents(sinceSeq int64) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	events, err := c.store.ListEventsSince(sinceSeq)
	if err != nil {
		return errorJSON(fmt.Sprintf("list events: %v", err))
	}
	out := make([]EventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, eventToDTO(event))
	}
	data, err := json.Marshal(out)
	if err != nil {
		return errorJSON(fmt.Sprintf("encode events: %v", err))
	}
	return string(data)
}

// ImportEvents accepts JSON-encoded EventDTO list.
func (c *Core) ImportEvents(eventsJSON string) string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	if c.keys == nil || !c.keys.IsUnlocked() {
		return errorJSON("keys not unlocked")
	}
	var items []EventDTO
	if err := json.Unmarshal([]byte(eventsJSON), &items); err != nil {
		return errorJSON(fmt.Sprintf("decode events: %v", err))
	}
	events := make([]model.Event, 0, len(items))
	for _, item := range items {
		event, err := dtoToEvent(item)
		if err != nil {
			return errorJSON(fmt.Sprintf("convert event: %v", err))
		}
		events = append(events, event)
	}
	events = dedupeEvents(events)
	sortEvents(events)
	if err := c.applyImportedEvents(events); err != nil {
		return errorJSON(fmt.Sprintf("apply events: %v", err))
	}
	if err := c.store.AppendEvents(events); err != nil {
		return errorJSON(fmt.Sprintf("append events: %v", err))
	}
	return ""
}

// GetSyncState returns JSON-encoded sync state.
func (c *Core) GetSyncState() string {
	if c.store == nil {
		return errorJSON("storage not initialized")
	}
	state, err := c.store.GetSyncState()
	if err != nil {
		return errorJSON(fmt.Sprintf("get sync state: %v", err))
	}
	if state.DeviceID == "" {
		state.DeviceID = c.deviceID
		_ = c.store.SaveSyncState(state)
	}
	data, err := json.Marshal(syncStateToDTO(state))
	if err != nil {
		return errorJSON(fmt.Sprintf("encode sync state: %v", err))
	}
	return string(data)
}

// InitKeys initializes encryption keys.
func (c *Core) InitKeys(passphrase string) string {
	if c.store == nil || c.keys == nil {
		return errorJSON("storage not initialized")
	}
	if passphrase == "" {
		return errorJSON("passphrase is required")
	}
	existing, err := c.store.GetKeyState()
	if err != nil {
		return errorJSON(fmt.Sprintf("get key state: %v", err))
	}
	if len(existing.Salt) > 0 {
		return errorJSON("keys already initialized")
	}
	salt, err := crypto.NewSalt()
	if err != nil {
		return errorJSON(fmt.Sprintf("salt: %v", err))
	}
	if err := c.keys.DeriveKey(passphrase, salt); err != nil {
		return errorJSON(fmt.Sprintf("derive key: %v", err))
	}
	state := model.KeyState{
		Salt:      salt,
		KDF:       "scrypt",
		UpdatedAt: time.Now().UTC(),
	}
	if err := c.store.SaveKeyState(state); err != nil {
		return errorJSON(fmt.Sprintf("save key state: %v", err))
	}
	return ""
}

// UnlockKeys unlocks encryption keys.
func (c *Core) UnlockKeys(passphrase string) string {
	if c.store == nil || c.keys == nil {
		return errorJSON("storage not initialized")
	}
	if passphrase == "" {
		return errorJSON("passphrase is required")
	}
	state, err := c.store.GetKeyState()
	if err != nil {
		return errorJSON(fmt.Sprintf("get key state: %v", err))
	}
	if len(state.Salt) == 0 {
		return errorJSON("keys not initialized")
	}
	if err := c.keys.DeriveKey(passphrase, state.Salt); err != nil {
		return errorJSON(fmt.Sprintf("derive key: %v", err))
	}
	return ""
}

// RotateKeys rotates encryption keys.
func (c *Core) RotateKeys() string { return "" }

// ReorderItemDTO represents a reorder mutation.
type ReorderItemDTO struct {
	ID      string `json:"id"`
	Order   int64  `json:"order"`
	DueDate string `json:"due_date"`
}

// SyncStateDTO is a bind-safe sync state.
type SyncStateDTO struct {
	LastSeq   int64  `json:"last_seq"`
	LastSync  string `json:"last_sync"`
	DeviceID  string `json:"device_id"`
	ServerTag string `json:"server_tag"`
}

func taskToDTO(task model.Task) TaskDTO {
	return TaskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     formatDate(task.DueDate),
		Order:       task.Order,
		CreatedAt:   formatTime(task.CreatedAt),
		UpdatedAt:   formatTime(task.UpdatedAt),
		CompletedAt: formatTime(task.CompletedAt),
		Archived:    task.Archived,
	}
}

func dtoToTask(dto TaskDTO) (model.Task, error) {
	dueDate, err := parseDate(dto.DueDate)
	if err != nil {
		return model.Task{}, err
	}
	createdAt, err := parseTime(dto.CreatedAt)
	if err != nil {
		return model.Task{}, err
	}
	updatedAt, err := parseTime(dto.UpdatedAt)
	if err != nil {
		return model.Task{}, err
	}
	completedAt, err := parseTime(dto.CompletedAt)
	if err != nil {
		return model.Task{}, err
	}
	return model.Task{
		ID:          dto.ID,
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
		Priority:    dto.Priority,
		DueDate:     dueDate,
		Order:       dto.Order,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		CompletedAt: completedAt,
		Archived:    dto.Archived,
	}, nil
}

func eventToDTO(event model.Event) EventDTO {
	encoded := ""
	if len(event.Payload) > 0 {
		encoded = base64.StdEncoding.EncodeToString(event.Payload)
	}
	return EventDTO{
		ID:          event.ID,
		DeviceID:    event.DeviceID,
		Seq:         event.Seq,
		TS:          formatTime(event.TS),
		Type:        event.Type,
		PayloadJSON: encoded,
	}
}

func dtoToEvent(dto EventDTO) (model.Event, error) {
	ts, err := parseTime(dto.TS)
	if err != nil {
		return model.Event{}, err
	}
	var payload []byte
	if dto.PayloadJSON != "" {
		decoded, err := base64.StdEncoding.DecodeString(dto.PayloadJSON)
		if err != nil {
			return model.Event{}, err
		}
		payload = decoded
	}
	return model.Event{
		ID:       dto.ID,
		DeviceID: dto.DeviceID,
		Seq:      dto.Seq,
		TS:       ts,
		Type:     dto.Type,
		Payload:  payload,
	}, nil
}

func syncStateToDTO(state model.SyncState) SyncStateDTO {
	return SyncStateDTO{
		LastSeq:   state.LastSeq,
		LastSync: formatTime(state.LastSync),
		DeviceID: state.DeviceID,
		ServerTag: state.ServerTag,
	}
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

func errorJSON(message string) string {
	out, _ := json.Marshal(map[string]string{"error": message})
	return string(out)
}

func (c *Core) appendEvent(eventType string, task model.Task) error {
	if c.store == nil {
		return fmt.Errorf("storage not initialized")
	}
	if c.keys == nil || !c.keys.IsUnlocked() {
		return fmt.Errorf("keys not unlocked")
	}
	plaintext, err := json.Marshal(taskToDTO(task))
	if err != nil {
		return fmt.Errorf("encode payload: %w", err)
	}
	payload, err := c.keys.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encrypt payload: %w", err)
	}
	state, err := c.store.GetSyncState()
	if err != nil {
		return fmt.Errorf("get sync state: %w", err)
	}
	if state.DeviceID == "" {
		state.DeviceID = c.deviceID
	}
	state.LastSeq++
	event := model.Event{
		ID:       uuid.NewString(),
		DeviceID: state.DeviceID,
		Seq:      state.LastSeq,
		TS:       time.Now().UTC(),
		Type:     eventType,
		Payload:  payload,
	}
	if err := c.store.AppendEvents([]model.Event{event}); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	if err := c.store.SaveSyncState(state); err != nil {
		return fmt.Errorf("save sync state: %w", err)
	}
	return nil
}

func (c *Core) applyImportedEvents(events []model.Event) error {
	for _, event := range events {
		exists, err := c.store.HasEvent(event.ID)
		if err != nil {
			return fmt.Errorf("check event: %w", err)
		}
		if exists {
			continue
		}
		plaintext, err := c.keys.Decrypt(event.Payload)
		if err != nil {
			return fmt.Errorf("decrypt event payload: %w", err)
		}
		event.Payload = plaintext
		taskID := taskIDFromPayload(event.Payload)
		if taskID == "" {
			return fmt.Errorf("missing task id in payload")
		}
		task, err := c.store.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}
		updated, changed, conflict, err := sync.ApplyEvent(task, event)
		if err != nil {
			return fmt.Errorf("apply event: %w", err)
		}
		if changed {
			if err := c.store.UpsertTask(updated); err != nil {
				return fmt.Errorf("upsert task: %w", err)
			}
			continue
		}
		if conflict {
			conflictRecord := model.Conflict{
				ID:             uuid.NewString(),
				TaskID:         task.ID,
				LocalUpdatedAt: task.UpdatedAt,
				RemoteEventID:  event.ID,
				RemoteTS:       event.TS,
				DetectedAt:     time.Now().UTC(),
				Resolution:     "lww_local",
			}
			if err := c.store.AddConflict(conflictRecord); err != nil {
				return fmt.Errorf("add conflict: %w", err)
			}
		}
	}
	return nil
}

func taskIDFromPayload(payload []byte) string {
	var dto TaskDTO
	_ = json.Unmarshal(payload, &dto)
	return dto.ID
}

func dedupeEvents(events []model.Event) []model.Event {
	seen := make(map[string]struct{}, len(events))
	out := make([]model.Event, 0, len(events))
	for _, event := range events {
		if event.ID == "" {
			continue
		}
		if _, ok := seen[event.ID]; ok {
			continue
		}
		seen[event.ID] = struct{}{}
		out = append(out, event)
	}
	return out
}

func sortEvents(events []model.Event) {
	sort.SliceStable(events, func(i, j int) bool {
		a := events[i]
		b := events[j]
		if a.DeviceID != b.DeviceID {
			return a.DeviceID < b.DeviceID
		}
		if a.Seq != b.Seq {
			return a.Seq < b.Seq
		}
		return a.TS.Before(b.TS)
	})
}

// DebugDecryptEvent is a local-only helper to decrypt an event payload.
func (c *Core) DebugDecryptEvent(payloadBase64 string) string {
	if c.keys == nil || !c.keys.IsUnlocked() {
		return errorJSON("keys not unlocked")
	}
	if payloadBase64 == "" {
		return errorJSON("payload is required")
	}
	raw, err := base64.StdEncoding.DecodeString(payloadBase64)
	if err != nil {
		return errorJSON(fmt.Sprintf("decode payload: %v", err))
	}
	plaintext, err := c.keys.Decrypt(raw)
	if err != nil {
		return errorJSON(fmt.Sprintf("decrypt payload: %v", err))
	}
	return string(plaintext)
}
