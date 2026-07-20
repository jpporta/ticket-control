package endofday

import (
	"context"
	"time"

	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/printer"
)

type Repo interface {
	CountUserTasksInWindow(ctx context.Context, createdBy int32, start, end time.Time) (int64, error)
	CompleteTasks(ctx context.Context, ids []int32, completedBy int32) (int64, error)
	CountUserCompletedTasksInWindow(ctx context.Context, completedBy int32, start, end time.Time) (int64, error)
}

type Printer interface {
	PrintEndOfDay(input printer.EndOfDayInput) error
}

type Service struct {
	repo    Repo
	printer Printer
}

func New(repo Repo, p Printer) *Service {
	return &Service{repo: repo, printer: p}
}

func (s *Service) EndOfDay(ctx context.Context, userID int32, userName string, noDone int, offset int) error {
	start, end := dayWindow(offset)
	tasks, err := s.repo.CountUserTasksInWindow(ctx, userID, start, end)
	if err != nil {
		return err
	}
	return s.printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       start,
		NoTasks:   int(tasks),
		NoDone:    noDone,
	})
}

func (s *Service) EndOfDayAuto(ctx context.Context, userID int32, userName string, offset int) error {
	start, end := dayWindow(offset)
	tasks, err := s.repo.CountUserTasksInWindow(ctx, userID, start, end)
	if err != nil {
		return err
	}
	done, err := s.repo.CountUserCompletedTasksInWindow(ctx, userID, start, end)
	if err != nil {
		return err
	}
	return s.printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       start,
		NoTasks:   int(tasks),
		NoDone:    int(done),
	})
}

func (s *Service) EndOfWeekend(ctx context.Context, userID int32, userName string, start, end time.Time) error {
	tasks, err := s.repo.CountUserTasksInWindow(ctx, userID, start, end)
	if err != nil {
		return err
	}
	done, err := s.repo.CountUserCompletedTasksInWindow(ctx, userID, start, end)
	if err != nil {
		return err
	}
	return s.printer.PrintEndOfDay(printer.EndOfDayInput{
		CreatedBy: userName,
		Day:       start,
		EndDay:    end,
		NoTasks:   int(tasks),
		NoDone:    int(done),
	})
}

func dayWindow(offset int) (time.Time, time.Time) {
	t := clock.Now().AddDate(0, 0, offset)
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return start, start.Add(24 * time.Hour)
}
