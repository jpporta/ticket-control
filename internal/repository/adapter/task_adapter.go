package adapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
	"github.com/jpporta/ticket-control/internal/task"
)

func (a Task) CreateTask(ctx context.Context, title, description string, priority int32, createdBy int32) (int32, error) {
	return a.Q.CreateTask(ctx, repository.CreateTaskParams{
		Title:       title,
		Description: pgtype.Text{String: description, Valid: description != ""},
		Priority:    pgtype.Int4{Int32: priority, Valid: priority > 0 && priority <= 5},
		CreatedBy:   createdBy,
	})
}

func (a Task) DeleteLastTask(ctx context.Context, createdBy int32) error {
	return a.Q.DeleteLastTask(ctx, createdBy)
}

func (a Task) GetOpenTasks(ctx context.Context, createdBy int32, limit, offset int32) ([]task.OpenTask, error) {
	rows, err := a.Q.GetOpenTasks(ctx, repository.GetOpenTasksParams{
		Limit:     limit,
		Offset:    offset,
		CreatedBy: createdBy,
	})
	if err != nil {
		return nil, err
	}
	out := make([]task.OpenTask, 0, len(rows))
	for _, r := range rows {
		out = append(out, task.OpenTask{
			ID:        r.ID,
			Title:     r.Title,
			Priority:  int(r.Priority.Int32),
			CreatedAt: r.CreatedAt.Time,
		})
	}
	return out, nil
}

func (a Task) CountUserTasksInWindow(ctx context.Context, createdBy int32, start, end time.Time) (int64, error) {
	return a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   createdBy,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
}

func (a Task) GetUserById(ctx context.Context, id int32) (string, error) {
	row, err := a.Q.GetUserById(ctx, id)
	if err != nil {
		return "", err
	}
	return row.Name, nil
}

func (a Task) MarkTaskAsDone(ctx context.Context, id int32, completedBy int32) error {
	return a.Q.MarkTaskAsDone(ctx, repository.MarkTaskAsDoneParams{
		ID:          id,
		CompletedBy: pgtype.Int4{Int32: completedBy, Valid: true},
	})
}
