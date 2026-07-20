package task

import (
	"context"
	"time"
)

// Repo is the slice of repository operations the task service needs.
// The concrete adapter lives in internal/repository/adapter and converts
// between plain Go types and the sqlc-generated pgtype wrappers.
type Repo interface {
	CreateTask(ctx context.Context, title, description string, priority int32, createdBy int32) (int32, error)
	DeleteLastTask(ctx context.Context, createdBy int32) error
	GetOpenTasks(ctx context.Context, createdBy int32, limit, offset int32) ([]OpenTask, error)
	CountUserTasksInWindow(ctx context.Context, createdBy int32, start, end time.Time) (int64, error)
	GetUserById(ctx context.Context, id int32) (string, error)
	MarkTaskAsDone(ctx context.Context, id int32, completedBy int32) error
}

// TaskInput is the printer-facing shape, used by the batch printer path.
type TaskInput struct {
	ID                 int32
	Title, Description string
	Priority           int32
	CreatedBy          string
	CreatedAt          time.Time
}

// CreateParams is what callers pass to Service.CreateTask.
type CreateParams struct {
	Title, Description string
	Priority           int32
}

// OpenTask is the JSON shape returned to clients.
type OpenTask struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Priority  int        `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}
