package schedule

import (
	"context"
	"time"
)

type Repo interface {
	CreateScheduleTask(ctx context.Context, name, title, description, cronExpression, checkFunction string, createdBy int32) (int32, error)
	ToggleScheduleTask(ctx context.Context, id, createdBy int32) (ToggleScheduleRow, error)
	GetUserScheduleTasks(ctx context.Context, createdBy int32) ([]ScheduleRow, error)
	GetAllEnabledScheduleTasks(ctx context.Context) ([]EnabledScheduleRow, error)
}

type ScheduleRow struct {
	ID             int32
	Name           string
	Title          string
	Description    string
	CronExpression string
	Enabled        bool
	CreatedAt      time.Time
	CheckFunction  string
}

type ToggleScheduleRow struct {
	ID             int32
	Name           string
	Title          string
	Description    string
	CronExpression string
	Enabled        bool
	CreatedBy      int32
	CheckFunction  string
}

type EnabledScheduleRow struct {
	ID             int32
	Name           string
	Title          string
	Description    string
	CronExpression string
	CreatedBy      int32
	CheckFunction  string
}

type CreateParams struct {
	Name           string
	Title          string
	Description    string
	CronExpression string
	CheckFunction  string
}

type Response struct {
	ID             int32     `json:"id"`
	Name           string    `json:"name"`
	Title          string    `json:"title"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	CronExpression string    `json:"cron_expression"`
	CheckFunction  string    `json:"check_function"`
	NextRun        time.Time `json:"next_run"`
	LastRun        time.Time `json:"last_run"`
}
