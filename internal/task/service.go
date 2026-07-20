package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/ports"
	"github.com/jpporta/ticket-control/internal/printer"
)

// Service holds the task use-cases. It depends on the Repo interface and the
// Printer port, not on the sqlc-generated types or the concrete *printer.Printer.
type Service struct {
	repo    Repo
	printer ports.Printer
}

func New(repo Repo, p ports.Printer) *Service {
	return &Service{repo: repo, printer: p}
}

// UserHasReachedTaskLimit reports whether the user has hit the daily task cap.
func (s *Service) UserHasReachedTaskLimit(ctx context.Context, userID int32) (bool, error) {
	start := clock.Today()
	total, err := s.repo.CountUserTasksInWindow(ctx, userID, start, start.Add(24*60*60*1e9))
	if err != nil {
		return false, fmt.Errorf("count user tasks: %w", err)
	}
	return total >= TaskLimit, nil
}

// CreateTask inserts a task row, prints it, and rolls back the row if printing fails.
func (s *Service) CreateTask(ctx context.Context, p CreateParams, userID int32) (int32, error) {
	limitReached, err := s.UserHasReachedTaskLimit(ctx, userID)
	if err != nil {
		return 0, err
	}
	if limitReached {
		return 0, fmt.Errorf("task: %w", apperr.ErrQuotaExceeded)
	}

	id, err := s.repo.CreateTask(ctx, p.Title, p.Description, p.Priority, userID)
	if err != nil {
		return 0, fmt.Errorf("create task: %w", err)
	}

	userName, err := s.repo.GetUserById(ctx, userID)
	if err != nil {
		_ = s.repo.DeleteLastTask(ctx, userID)
		return 0, fmt.Errorf("lookup user: %w", err)
	}

	if err := s.printer.PrintTask(id, p.Title, p.Description, p.Priority, userName, clock.Now()); err != nil {
		_ = s.repo.DeleteLastTask(ctx, userID)
		if errors.Is(err, errPrinterOffline) {
			return 0, fmt.Errorf("print task: %w", apperr.ErrPrinterOffline)
		}
		return 0, fmt.Errorf("print task: %w", err)
	}
	return id, nil
}

// MarkTaskAsDone completes the task and emits the audible "bip".
func (s *Service) MarkTaskAsDone(ctx context.Context, taskID, userID int32) error {
	if err := s.repo.MarkTaskAsDone(ctx, taskID, userID); err != nil {
		return fmt.Errorf("mark task done: %w", err)
	}
	return s.printer.PrintBip()
}

// GetOpenTasks returns the user's open tasks, paginated.
func (s *Service) GetOpenTasks(ctx context.Context, userID int32, amount, page int) ([]OpenTask, error) {
	rows, err := s.repo.GetOpenTasks(ctx, userID, int32(amount), int32(page*amount))
	if err != nil {
		return nil, fmt.Errorf("get open tasks: %w", err)
	}
	out := make([]OpenTask, 0, len(rows))
	for _, r := range rows {
		out = append(out, OpenTask{
			ID:        r.ID,
			Title:     r.Title,
			Priority:  int(r.Priority),
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}

// CreateTasks is the batch path used by the events endpoint.
// It does NOT apply a quota check — that happens upstream in events.Service.
func (s *Service) CreateTasks(ctx context.Context, tasks []CreateParams, userID int32) (int, error) {
	printerInputs := make([]printer.TaskInput, 0, len(tasks))
	for _, t := range tasks {
		id, err := s.repo.CreateTask(ctx, t.Title, t.Description, t.Priority, userID)
		if err != nil {
			continue
		}
		userName, err := s.repo.GetUserById(ctx, userID)
		if err != nil {
			_ = s.repo.DeleteLastTask(ctx, userID)
			continue
		}
		printerInputs = append(printerInputs, printer.TaskInput{
			ID:          id,
			Title:       t.Title,
			Description: t.Description,
			Priority:    t.Priority,
			CreatedBy:   userName,
			CreatedAt:   clock.Now(),
		})
	}
	if err := s.printer.PrintTasks(printerInputs); err != nil {
		return 0, fmt.Errorf("print tasks: %w", err)
	}
	return len(printerInputs), nil
}

// errPrinterOffline is the unexported sentinel returned by the printer package
// when the device is disabled. Re-declared here to avoid an import cycle.
var errPrinterOffline = errors.New("printer offline")
