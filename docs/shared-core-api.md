# Shared Go Core API (Draft)

Goal: keep all app logic in a single Go core, and expose a bind-safe API for native UIs on iOS, macOS, and Windows.

## Principles
- Core logic lives in Go: data model, validation, encryption, sync, conflict resolution.
- Bind-safe API only: no interfaces, channels, funcs, maps, or `time.Time` across the boundary.
- Use DTOs with primitive fields and RFC3339/ISO date strings.
- Keep persistence behind a Go interface with platform-specific Go implementations.

## Package Layout (Proposed)
- `core/`
  - `core/model/` (internal types with `time.Time`)
  - `core/logic/` (validation, ordering, merge, conflict resolution)
  - `core/crypto/` (E2EE)
  - `core/sync/` (event log, replay, merge)
  - `core/storage/` (interfaces)
  - `core/storage/sqlite/` (default SQLite implementation)
  - `core/platform/ios/` (build-tagged storage/helpers)
  - `core/platform/desktop/` (build-tagged storage/helpers)
- `core/bind/` (bind-safe API surface)

## Bind-Safe DTOs
All fields are primitives or slices of primitives.

```
TaskDTO {
  id: string
  title: string
  description: string
  status: string        // "active" | "done"
  priority: string      // "low" | "med" | "high"
  due_date: string      // "YYYY-MM-DD" or ""
  order: int64
  created_at: string    // RFC3339
  updated_at: string    // RFC3339
  completed_at: string  // RFC3339 or ""
  archived: bool
}
```

```
EventDTO {
  id: string
  device_id: string
  seq: int64
  ts: string           // RFC3339
  type: string         // "create" | "update" | "delete" | "reorder" | ...
  payload_json: string // encrypted or plaintext depending on layer
}
```

## Public Bind API (Go)
Expose a small facade with JSON in/out to maximize bind compatibility.

```
type Core struct

// Lifecycle
func NewCore(configJSON string) *Core
func (c *Core) Open() string                  // "" on success, error string otherwise
func (c *Core) Close() string                 // "" on success, error string otherwise

// Tasks
func (c *Core) ListTasks(filterJSON string) string
func (c *Core) CreateTask(taskJSON string) string
func (c *Core) UpdateTask(taskJSON string) string
func (c *Core) DeleteTask(taskID string) string
func (c *Core) ReorderTasks(reorderJSON string) string
func (c *Core) SetDueDate(taskID string, dueDate string) string
func (c *Core) SetCompleted(taskID string, completed bool) string

// Sync
func (c *Core) ExportEvents(sinceSeq int64) string
func (c *Core) ImportEvents(eventsJSON string) string
func (c *Core) GetSyncState() string

// Keys / Encryption
func (c *Core) InitKeys(passphrase string) string
func (c *Core) UnlockKeys(passphrase string) string
func (c *Core) RotateKeys() string
```

Notes:
- Return values are JSON strings or empty string for success + error string on failure.
- This avoids bind limitations and makes Swift/Windows interop straightforward.
- The bind layer converts JSON DTOs into internal `core/model` types.

## Storage Interface (Go-only)
Interfaces stay inside Go. Platform-specific Go files implement storage with build tags.

```
type Storage interface {
  Open() error
  Close() error

  ListTasks(filter TaskFilter) ([]model.Task, error)
  UpsertTask(task model.Task) error
  DeleteTask(id string) error

  AppendEvents(events []model.Event) error
  ListEventsSince(seq int64) ([]model.Event, error)
  GetSyncState() (model.SyncState, error)
  SaveSyncState(state model.SyncState) error
}
```

## Bindability Constraints
- No `time.Time` across the boundary; always encode as string.
- No interfaces or generics in public API.
- Avoid maps in public API; use JSON strings instead.
- Keep API surface stable and versioned.

## Immediate Implementation Steps
- Create `core/` module and `core/bind/` facade.
- Replace `time.Time` in bind DTOs with string fields.
- Create JSON encoding helpers in `core/bind`.
- Implement a minimal `Storage` SQLite backend in Go with build tags for desktop and iOS.
