package internal

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Application struct {
	Q     *repository.Queries
	Cron  *CronJob
	Close func() error
}

func New(conn *pgxpool.Pool) *Application {
	cronJob := NewCronJob()
	return &Application{
		Q:    repository.New(conn),
		Cron: cronJob,
	}
}
