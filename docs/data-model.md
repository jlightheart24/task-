# Data Model (Draft)

## Task (Encrypted Payload)
Encrypted JSON payload stored in local SQLite and sent to server.

Fields (encrypted):
- id
- title
- status (open|done)
- priority
- due_date (optional)
- created_at
- updated_at
- completed_at (optional)
- recurrence (daily|eod|none)
- archived (bool)

## Local SQLite Tables
- tasks
  - id (uuid)
  - ciphertext (blob)
  - created_at (int64)
  - updated_at (int64)
  - deleted_at (int64, nullable)
  - version (int64)

- task_events
  - event_id (uuid)
  - task_id (uuid)
  - op (text)
  - timestamp (int64)
  - device_id (uuid)
  - ciphertext (blob)
  - synced (bool)

- conflicts
  - id (uuid)
  - task_id (uuid)
  - local_event_id (uuid)
  - remote_event_id (uuid)
  - detected_at (int64)
  - resolved_at (int64, nullable)
  - resolution (text)

- devices
  - id (uuid)
  - name (text)
  - created_at (int64)

- settings
  - key (text, primary key)
  - value (text)

## Server Postgres Tables
- users
  - id (uuid)
  - username (text, unique)
  - password_hash (text)
  - created_at (timestamp)

- devices
  - id (uuid)
  - user_id (uuid)
  - name (text)
  - created_at (timestamp)

- task_events
  - event_id (uuid)
  - user_id (uuid)
  - task_id (uuid)
  - op (text)
  - timestamp (timestamp)
  - device_id (uuid)
  - ciphertext (bytea)

- recovery_wrapped_dek
  - user_id (uuid, primary key)
  - dek_wrapped (bytea)
  - created_at (timestamp)
