package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Event struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Title string    `json:"title"`
}

func (a *Application) CreateEvents(ctx context.Context, events []Event, userId int32) (int, error) {
	applicationTasks := []CreateTaskParams{}
	hasUserPrintedToday, err := a.hasUserAlreadyPrintedToday(ctx, userId)
	if err != nil {
		return 0, fmt.Errorf("Error checking if user has printed today: %w", err)
	}
	if hasUserPrintedToday {
		return 0, fmt.Errorf("User has already printed today, cannot create events")
	}
	for _, event := range events {
		description := event.Start.Format("15:04") + " - " + event.End.Format("15:04")
		applicationTasks = append(applicationTasks, CreateTaskParams{
			Title:       event.Title,
			Description: description,
			Priority:    -1,
		})
	}
	noCreated, err := a.CreateTasks(ctx, applicationTasks, userId)
	if err != nil {
		return 0, fmt.Errorf("Error creating events: %w", err)
	}
	return noCreated, nil
}

func (a *Application) hasUserAlreadyPrintedToday(ctx context.Context, userId int32) (bool, error) {
	t := time.Now()
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := start.Add(24 * time.Hour)
	noPrintedToday, err := a.Q.GetAccessStats(ctx, repository.GetAccessStatsParams{
		UserID:    pgtype.Int4{Int32: userId, Valid: userId != 0},
		AccessedAt: pgtype.Timestamp{Time: start, Valid: true},
		AccessedAt_2: pgtype.Timestamp{Time: end, Valid: true},
		Path:      "/events",
		Method:    "POST",
	})
	if err != nil {
		return false, fmt.Errorf("Error checking if user has printed today: %w", err)
	}
	return noPrintedToday > 0, nil
}
