package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

var TASK_LIMIT int64 = 50

func (a *Application) UserHasReachedTaskLimit(ctx context.Context, userId int32) (bool, error) {
	startYear, startMonth, startDay := time.Now().Date()
	startTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)

	total, err := a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: startTime.Add(time.Hour * 24), Valid: true},
	})
	if err != nil {
		return false, err
	}

	return (total >= TASK_LIMIT), nil
}

func (a *Application) CreateTask(ctx context.Context, title, description string, priority int32, userId int32) (int32, error) {
	// Create in DB
	res, err := a.Q.CreateTask(ctx, repository.CreateTaskParams{
		Title:       title,
		Description: pgtype.Text{String: description, Valid: description != ""},
		Priority:    pgtype.Int4{Int32: priority, Valid: priority > 0 && priority <= 5},
		CreatedBy:   userId,
	})
	if err != nil {
		return 0, fmt.Errorf("Error creating task")
	}

	// Print, and if it fails, delete from DB
	p := printer.New(ctx)
	close, err := p.Start()
	if err != nil {
		err := a.Q.DeleteLastTask(ctx, userId)
		if err != nil {
			return 0, fmt.Errorf("Error deleting task after printer start failure: %w", err)
		}
		return 0, fmt.Errorf("Error starting printer: %w", err)
	}
	defer close()
	name := ctx.Value("userName").(string)
	err = p.PrintTask(title, description, priority, name)
	return res, nil
}
