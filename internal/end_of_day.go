package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a *Application) EndOfDay(ctx context.Context, userId int32, userName string, noDone int) error {
	t := time.Now()
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := start.Add(24 * time.Hour)
	taskCreatedToday, err := a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("Error getting tasks created today: %w", err)
	}
	return a.Printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       time.Now(),
		NoTasks:   int(taskCreatedToday),
		NoDone:    noDone,
	})
}
