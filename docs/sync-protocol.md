# Sync Protocol (Draft)

## Principles
- Offline-first: local DB is always authoritative while offline.
- Event-based sync: mutations logged as task_events with monotonically increasing sequence numbers.
- LWW conflict resolution with user notifications (until richer merge rules are defined).
- Server stores encrypted payloads only (no plaintext).

## Data Model (Events)
Each event captures a single mutation.

Event fields:
- id: UUID
- device_id: UUID (stable per device)
- seq: int64 (monotonic, local sequence)
- ts: RFC3339 timestamp (event creation time)
- type: string (`create`, `update`, `delete`, `reorder`, `set_due_date`, `set_completed`)
- payload: encrypted JSON blob (TaskDTO), base64-encoded for transport

## Sync State
Per client:
- last_seq: last local sequence number
- last_sync: last successful sync time
- device_id: stable ID
- server_tag: server cursor/etag (optional)

## Client Flow
1. Write to local SQLite immediately.
2. Log task_event with encrypted payload (append-only).
3. When online:
   - POST local events since `last_seq` to server.
   - GET remote events since server cursor (or timestamp).
4. Apply remote events locally in seq order per device.
5. Detect conflicts and log in conflicts table.

## API Endpoints (Draft)
- POST /auth/signup
- POST /auth/login
- POST /auth/reset-password
- POST /sync/events
- GET /sync/events?since=...

## Conflict Handling
- LWW applied automatically.
- Conflicts recorded locally with references to local and remote events.
- Notification level controlled by user settings:
  - none
  - summary
  - immediate

## Notes
- Events are idempotent by `id`; server should de-dup by event id.
- Clients must preserve event order per device (`seq`).
