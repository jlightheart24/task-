sequenceDiagram
  autonumber
  participant App as Flutter App (Device B)
  participant Local as Local DB (Isar/Drift)
  participant API as NestJS API
  participant DB as PostgreSQL
  participant Push as APNs/WNS

  Note over API,DB: Another device (A) already wrote a change

  App->>Local: User edits task (offline OK)
  App->>API: POST /tasks (JWT) with client_id, updated_at
  API->>DB: Transaction: upsert task, assign global version
  DB-->>API: {task, version}
  API-->>App: 200 OK {task with server version}
  API->>Push: Enqueue silent push for project members

  Push-->>App: Silent push (content-available)
  App->>Local: Wake sync worker
  App->>API: GET /sync?since=last_version
  API->>DB: Fetch rows where version > since (incl. tombstones/soft deletes)
  DB-->>API: Changed rows + next_version
  API-->>App: {tasks, projects, tombstones, next_version}
  App->>Local: Upsert/merge rows, apply deletions
  App->>App: Update last_version = next_version
