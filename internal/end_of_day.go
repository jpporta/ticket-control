package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a *Application) getUserTasksOfToday(ctx context.Context, userId int32, offset int) (int64, error) {
	t := time.Now()
	t = t.AddDate(0, 0, offset)
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := start.Add(24 * time.Hour)
	taskCreatedToday, err := a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("Error getting tasks created today: %w", err)
	}
	return taskCreatedToday, nil
}

func (a *Application) EndOfDay(ctx context.Context, userId int32, userName string, noDone int, offset int) error {
	taskCreatedToday, err := a.getUserTasksOfToday(ctx, userId, offset)
	if err != nil {
		return fmt.Errorf("Error getting tasks created today: %w", err)
	}
	return a.Printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       time.Now().AddDate(0, 0, offset),
		NoTasks:   int(taskCreatedToday),
		NoDone:    noDone,
	})
}
func (a *Application) EndOfDayWithTasks(ctx context.Context, userId int32, userName string, doneTasks []int32, offset int) error {
	taskCreatedToday, err := a.getUserTasksOfToday(ctx, userId, offset)
	if err != nil {
		return fmt.Errorf("Error getting tasks created today: %w", err)
	}
	noDone, err := a.Q.CompleteTasks(ctx, repository.CompleteTasksParams{
		Column1:   doneTasks,
		CreatedBy: userId,
	})

	if err != nil {
		return fmt.Errorf("Error completing tasks: %w", err)
	}
	return a.Printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       time.Now(),
		NoTasks:   int(taskCreatedToday),
		NoDone:    int(noDone),
	})
}

func (a *Application) EndOfDayAuto(ctx context.Context, userId int32, userName string, offset int) error {
	taskCreatedToday, err := a.getUserTasksOfToday(ctx, userId, offset)
	if err != nil {
		return fmt.Errorf("Error getting tasks created today: %w", err)
	}

	t := time.Now()
	t = t.AddDate(0, 0, offset)
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := start.Add(24 * time.Hour)

	noDoneToday, err := a.Q.GetNoCompletedTasks(ctx, repository.GetNoCompletedTasksParams{
		CompletedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CompletedAt_2: pgtype.Timestamp{Time: end, Valid: true},
		CreatedBy:     userId,
	})
	if err != nil {
		return fmt.Errorf("Error completing tasks: %w", err)
	}
	return a.Printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       t,
		NoTasks:   int(taskCreatedToday),
		NoDone:    int(noDoneToday),
	})
}
