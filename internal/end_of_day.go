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
	p := printer.New(ctx)
	close, err := p.Start()
	if err != nil {
		return fmt.Errorf("Error starting printer: %w", err)
	}
	defer close()
	taskCreatedToday, err := a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: time.Now().Truncate(24 * time.Hour), Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("Error getting tasks created today: %w", err)
	}
	return p.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       time.Now(),
		NoTasks:   int(taskCreatedToday),
		NoDone:    noDone,
	})
}
