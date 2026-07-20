package events

import (
	"context"
	"fmt"
	"time"

	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/task"
)

type Repo interface {
	GetAccessStats(ctx context.Context, userID int32, path, method string, start, end time.Time) (int64, error)
}

// Service is the Apple Shortcuts /events endpoint.
type Service struct {
	repo Repo
	task *task.Service
}

func New(repo Repo, t *task.Service) *Service {
	return &Service{repo: repo, task: t}
}

// CreateEvents imports a day of calendar events as tasks. It first guards
// against duplicate imports in the same UTC day.
func (s *Service) CreateEvents(ctx context.Context, events []Event, userID int32) (int, error) {
	already, err := s.UserPrintedSinceUTCMidnight(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("check duplicate: %w", err)
	}
	if already {
		return 0, fmt.Errorf("events: already imported today")
	}
	params := make([]task.CreateParams, 0, len(events))
	for _, e := range events {
		description := e.Start.Format("15:04") + " - " + e.End.Format("15:04")
		params = append(params, task.CreateParams{
			Title:       e.Title,
			Description: description,
			Priority:    -1,
		})
	}
	return s.task.CreateTasks(ctx, params, userID)
}

// UserPrintedSinceUTCMidnight reports whether the user has already imported
// events today. Uses an `> 1` threshold against the access log; off-by-one is
// intentional and matches the behaviour the existing shortcurts depend on.
func (s *Service) UserPrintedSinceUTCMidnight(ctx context.Context, userID int32) (bool, error) {
	start := clock.Today()
	end := start.Add(24 * time.Hour)
	no, err := s.repo.GetAccessStats(ctx, userID, "/events", "POST", start, end)
	if err != nil {
		return false, fmt.Errorf("get access stats: %w", err)
	}
	return no > 1, nil
}
