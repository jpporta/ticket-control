package adapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a EndOfDay) CountUserTasksInWindow(ctx context.Context, createdBy int32, start, end time.Time) (int64, error) {
	return a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   createdBy,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
}

func (a EndOfDay) CompleteTasks(ctx context.Context, ids []int32, completedBy int32) (int64, error) {
	return a.Q.CompleteTasks(ctx, repository.CompleteTasksParams{
		Column1:     ids,
		CompletedBy: pgtype.Int4{Int32: completedBy, Valid: true},
	})
}

func (a EndOfDay) CountUserCompletedTasksInWindow(ctx context.Context, completedBy int32, start, end time.Time) (int64, error) {
	return a.Q.GetNoCompletedTasks(ctx, repository.GetNoCompletedTasksParams{
		CompletedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CompletedAt_2: pgtype.Timestamp{Time: end, Valid: true},
		CompletedBy:   pgtype.Int4{Int32: completedBy, Valid: true},
	})
}
