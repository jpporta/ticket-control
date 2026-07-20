package schedule

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jpporta/ticket-control/internal/ports"
	"github.com/robfig/cron/v3"
)

// PossibleCheckFunctions lists the names a client may put in `check_function`.
var PossibleCheckFunctions = []string{
	"is_last_workday_of_month",
	"is_last_weekday_of_middle",
	"is_last_weekday_of_10",
}

// Job is the in-memory representation passed to the scheduler.
type Job struct {
	ID             int32
	Name           string
	Title          string
	Description    string
	CronExpression string
	CreatedBy      int32
	CheckFunction  string
}

// CheckFunc is the signature every registered predicate has. It receives a
// closure to invoke if the predicate fires.
type CheckFunc func(func())

// Scheduler abstracts robfig/cron so the schedule service doesn't depend on
// it directly. The default implementation lives in this package.
type Scheduler interface {
	Add(expr string, fn func()) (cron.EntryID, error)
	Remove(id cron.EntryID)
	Start()
	Stop()
	Entries() []cron.Entry
}

type Service struct {
	repo         Repo
	printer      ports.Printer
	scheduler    Scheduler
	checkFuncs   map[string]CheckFunc
	jobs         map[int32]cron.EntryID
	taskCallback func(ctx context.Context, title, description string, userID int32) (int32, error)
}

// New wires the schedule service. The taskCallback is invoked when a scheduled
// job fires; passing it in keeps the schedule package free of any import of
// the task package.
func New(repo Repo, p ports.Printer, taskCallback func(ctx context.Context, title, description string, userID int32) (int32, error), checkFuncs map[string]CheckFunc) *Service {
	return &Service{
		repo:         repo,
		printer:      p,
		scheduler:    &defaultScheduler{c: cron.New()},
		checkFuncs:   checkFuncs,
		jobs:         make(map[int32]cron.EntryID),
		taskCallback: taskCallback,
	}
}

// Start registers internal quiet-hours jobs and loads any user-defined
// schedules from the DB.
func (s *Service) Start(ctx context.Context) error {
	if _, err := s.scheduler.Add("0 22 * * *", func() { s.printer.Toggle(false) }); err != nil {
		return fmt.Errorf("add quiet-hours-on job: %w", err)
	}
	if _, err := s.scheduler.Add("0 8 * * *", func() { s.printer.Toggle(true) }); err != nil {
		return fmt.Errorf("add quiet-hours-off job: %w", err)
	}
	jobs, err := s.repo.GetAllEnabledScheduleTasks(ctx)
	if err != nil {
		return fmt.Errorf("load enabled schedules: %w", err)
	}
	for _, j := range jobs {
		s.scheduleJob(Job{
			ID:             j.ID,
			Name:           j.Name,
			Title:          j.Title,
			Description:    j.Description,
			CronExpression: j.CronExpression,
			CreatedBy:      j.CreatedBy,
			CheckFunction:  j.CheckFunction,
		})
	}
	s.scheduler.Start()
	return nil
}

func (s *Service) Stop() { s.scheduler.Stop() }

func (s *Service) CreateSchedule(ctx context.Context, p CreateParams, userID int32) error {
	id, err := s.repo.CreateScheduleTask(ctx, p.Name, p.Title, p.Description, p.CronExpression, p.CheckFunction, userID)
	if err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}
	s.scheduleJob(Job{
		ID:             id,
		Name:           p.Name,
		Title:          p.Title,
		Description:    p.Description,
		CronExpression: p.CronExpression,
		CreatedBy:      userID,
		CheckFunction:  p.CheckFunction,
	})
	return nil
}

func (s *Service) GetSchedules(ctx context.Context, userID int32) ([]Response, error) {
	rows, err := s.repo.GetUserScheduleTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user schedules: %w", err)
	}
	var out []Response
	for _, r := range rows {
		entryID, ok := s.jobs[r.ID]
		var nextRun, lastRun time.Time
		if ok {
			for _, e := range s.scheduler.Entries() {
				if e.ID == entryID {
					nextRun = e.Next
					lastRun = e.Prev
					break
				}
			}
		}
		out = append(out, Response{
			ID:             r.ID,
			Name:           r.Name,
			Title:          r.Title,
			Enabled:        r.Enabled,
			CreatedAt:      r.CreatedAt,
			CronExpression: r.CronExpression,
			CheckFunction:  r.CheckFunction,
			NextRun:        nextRun,
			LastRun:        lastRun,
		})
	}
	return out, nil
}

func (s *Service) ToggleSchedule(ctx context.Context, id, userID int32) error {
	row, err := s.repo.ToggleScheduleTask(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("toggle schedule: %w", err)
	}
	if row.Enabled {
		s.scheduleJob(Job{
			ID:             row.ID,
			Name:           row.Name,
			Title:          row.Title,
			Description:    row.Description,
			CronExpression: row.CronExpression,
			CreatedBy:      userID,
			CheckFunction:  row.CheckFunction,
		})
	} else {
		if entryID, ok := s.jobs[id]; ok {
			s.scheduler.Remove(entryID)
			delete(s.jobs, id)
		}
	}
	return nil
}

// scheduleJob registers a Job with the cron scheduler. The closure respects
// an optional check_function predicate.
func (s *Service) scheduleJob(j Job) {
	entryID, err := s.scheduler.Add(j.CronExpression, func() {
		run := func() {
			if s.taskCallback == nil {
				return
			}
			if _, err := s.taskCallback(context.Background(), j.Title, j.Description, j.CreatedBy); err != nil {
				slog.Error("scheduled task", "id", j.ID, "name", j.Name, "err", err)
			}
		}
		if j.CheckFunction == "" {
			run()
			return
		}
		if pred, ok := s.checkFuncs[j.CheckFunction]; ok {
			pred(run)
		} else {
			slog.Warn("unknown check function, running unconditionally", "name", j.CheckFunction)
			run()
		}
	})
	if err != nil {
		slog.Error("add job", "id", j.ID, "name", j.Name, "err", err)
		return
	}
	s.jobs[j.ID] = entryID
}

var _ Scheduler = (*defaultScheduler)(nil)

type defaultScheduler struct{ c *cron.Cron }

func (d *defaultScheduler) Add(expr string, fn func()) (cron.EntryID, error) {
	return d.c.AddFunc(expr, fn)
}
func (d *defaultScheduler) Remove(id cron.EntryID) { d.c.Remove(id) }
func (d *defaultScheduler) Start()                 { d.c.Start() }
func (d *defaultScheduler) Stop()                  { d.c.Stop() }
func (d *defaultScheduler) Entries() []cron.Entry  { return d.c.Entries() }

// IsValidCheckFunction reports whether name is a registered predicate.
func (s *Service) IsValidCheckFunction(name string) bool {
	if name == "" {
		return true
	}
	_, ok := s.checkFuncs[name]
	return ok
}

var ErrNoSchedule = errors.New("no schedule registered")
