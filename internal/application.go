package internal

import (
	"github.com/jackc/pgx/v5"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Application struct {
	Q     *repository.Queries
	Close func() error
}

func New(conn *pgx.Conn) *Application {
	return &Application{
		Q: repository.New(conn),
	}
}
