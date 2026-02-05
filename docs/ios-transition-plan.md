# iOS Transition Plan

This plan captures the agreed roadmap for transitioning from the current desktop app to a desktop + iOS setup, with shared data model, encryption, and sync.

## Phase 0: Foundations
- Align data model across desktop and iOS.
- Finalize sync/conflict specification.
- Finalize encryption scheme (E2EE) and key management flow.
- Choose iOS stack.
- Choose desktop stack (macOS + Windows).
- Define shared Go core boundary (bind-friendly API).
  - Keep all business logic in Go (data model, encryption, sync, conflict resolution).
  - Expose only bind-safe types (no interfaces, channels, or `time.Time`).
  - Use DTOs with strings/ints/arrays and RFC3339 date strings.
  - Keep persistence behind a simple interface implemented per platform.

## Phase 1: iOS MVP (Local-Only)
- Implement local-only CRUD on iOS (offline-first).
- Implement key management UX on iOS.

## Phase 2: Encryption + Sync
- Implement encryption on desktop client.
- Implement encryption on iOS client.
- Build sync server + authentication.
- Implement sync on both clients.

## Phase 3: UX Parity
- Bring iOS UX up to parity with desktop.
- Polish interaction and performance on both platforms.

## Phase 4: Hardening + Release
- Backup/restore flows.
- Security review and threat model validation.
- App Store preparation and release readiness.

## Immediate Next Steps
- Choose iOS stack.
- Lock data model and sync/conflict spec.
- Implement encryption on desktop.
- Decide desktop targets (macOS + Windows) and desktop UI stack.
- Design the shared Go core API surface for bindings.
