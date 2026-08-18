// Package service wires the domain services together with the Printer port
// and exposes the result as a typed Services struct. cmd/web/main.go is the
// only caller of NewServices; everything else takes a *Services.
package service

import (
	"context"

	"github.com/jpporta/ticket-control/internal/endofday"
	"github.com/jpporta/ticket-control/internal/events"
	"github.com/jpporta/ticket-control/internal/letter"
	"github.com/jpporta/ticket-control/internal/link"
	"github.com/jpporta/ticket-control/internal/list"
	"github.com/jpporta/ticket-control/internal/ports"
	"github.com/jpporta/ticket-control/internal/schedule"
	"github.com/jpporta/ticket-control/internal/task"
)

// Services bundles every domain service plus the printer port.
type Services struct {
	Task     *task.Service
	List     *list.Service
	Link     *link.Service
	Letter   *letter.Service
	Schedule *schedule.Service
	EndOfDay *endofday.Service
	Events   *events.Service
	Printer  ports.Printer
}

// Deps groups the per-domain repo adapters the container needs.
type Deps struct {
	Task     task.Repo
	List     list.Repo
	Link     link.Repo
	Letter   letter.Repo
	Schedule schedule.Repo
	EndOfDay endofday.Repo
	Events   events.Repo
}

// NewServices constructs the container and wires the schedule → task callback.
// The schedule → task callback uses context.Background because robfig/cron
// does not propagate a request context; the underlying task creation does
// its own clock-based quota check.
func NewServices(d Deps, printer ports.Printer, checkFuncs map[string]schedule.CheckFunc) *Services {
	t := task.New(d.Task, printer)
	l := list.New(d.List, printer)
	ln := link.New(d.Link, printer)
	let := letter.New(d.Letter, printer)
	sc := schedule.New(d.Schedule, printer,
		func(_ context.Context, title, description string, userID int32) (int32, error) {
			return t.CreateTask(context.Background(), task.CreateParams{
				Title:       title,
				Description: description,
				Priority:    0,
			}, userID)
		},
		checkFuncs,
	)
	eod := endofday.New(d.EndOfDay, printer)
	ev := events.New(d.Events, t)
	return &Services{
		Task:     t,
		List:     l,
		Link:     ln,
		Letter:   let,
		Schedule: sc,
		EndOfDay: eod,
		Events:   ev,
		Printer:  printer,
	}
}
