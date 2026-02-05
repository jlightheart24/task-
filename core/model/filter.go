package model

// TaskFilter is used for querying tasks.
type TaskFilter struct {
	Status   string
	Archived *bool
	DueDate  string
}
