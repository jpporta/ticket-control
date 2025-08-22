package internal

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Application struct {
	Q     *repository.Queries
	Cron  *CronJob
	Close func() error
	Printer *printer.Printer
}

func New(conn *pgxpool.Pool) *Application {
	cronJob := NewCronJob()
	printer := printer.New(context.Background())
	return &Application{
		Q:    repository.New(conn),
		Cron: cronJob,
		Printer: printer,
	}
}

