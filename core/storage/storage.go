package storage

import "taskpp/core/model"

// Storage abstracts persistence for the core.
type Storage interface {
	Open() error
	Close() error

	ListTasks(filter model.TaskFilter) ([]model.Task, error)
	GetTask(id string) (model.Task, error)
	UpsertTask(task model.Task) error
	DeleteTask(id string) error

	AppendEvents(events []model.Event) error
	ListEventsSince(seq int64) ([]model.Event, error)
	GetSyncState() (model.SyncState, error)
	SaveSyncState(state model.SyncState) error

	GetKeyState() (model.KeyState, error)
	SaveKeyState(state model.KeyState) error
}
