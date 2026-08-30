package model

import "time"

type TaskStatus string

const (
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task represents work accepted by GoFlow. Execution is introduced in a later phase.
type Task struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Payload   string     `json:"payload"`
	Status    TaskStatus `json:"status"`
	Priority  int        `json:"priority"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
