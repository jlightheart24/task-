# Architecture

## Overview
- Desktop-first app built with Wails (Go backend + React frontend).
- Offline-first client with local SQLite and in-memory indexes.
- Sync server stores encrypted event payloads only (E2EE).

## Components
- Desktop client
  - Go backend: data storage, encryption, sync, indexing.
  - React frontend: editor-style task entry, views, calendar.
- Sync server
  - Auth + device registration.
  - Event-based sync API.
  - No access to plaintext tasks.

## Guiding Principles
- Offline-first, seamless sync when online.
- End-to-end encryption by default.
- Minimal friction UX.
- Clear separation of concerns to allow later repo split.
