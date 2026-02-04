# Sync Protocol (Draft)

## Principles
- Offline-first: local DB is always authoritative while offline.
- Event-based sync: mutations logged as task_events.
- LWW conflict resolution with user notifications.

## Client Flow
1. Write to local SQLite immediately.
2. Log task_event with encrypted payload.
3. When online:
   - POST local events to server.
   - GET remote events since last sync timestamp.
4. Apply remote events locally.
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
