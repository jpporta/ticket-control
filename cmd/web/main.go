package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
	"github.com/jpporta/ticket-control/internal/repository/adapter"
	"github.com/jpporta/ticket-control/internal/schedule"
	"github.com/jpporta/ticket-control/internal/service"
	"github.com/jpporta/ticket-control/internal/utils"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func wrapper(h http.HandlerFunc, fs ...middleware) http.HandlerFunc {
	if len(fs) == 0 {
		return h
	}
	next := fs[0]
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r, wrapper(h, fs[1:]...))
	}
}

func chainMiddleware(fs ...middleware) func(h http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return wrapper(h, fs...)
	}
}

func main() {
	initLogger()

	ctx := context.Background()
	conn, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	q := repository.New(conn)
	p := printer.New(ctx)

	checkFuncs := map[string]schedule.CheckFunc{
		"is_last_workday_of_month":  utils.IsLastWeekdayMonth,
		"is_last_weekday_of_middle": utils.IsLastWorkdayToMiddle,
		"is_last_weekday_of_10":     utils.IsLastWorkdayTo10,
	}

	svcs := service.NewServices(service.Deps{
		Task:     adapter.Task{Q: q},
		List:     adapter.List{Q: q},
		Link:     adapter.Link{Q: q},
		Schedule: adapter.Schedule{Q: q},
		EndOfDay: adapter.EndOfDay{Q: q},
		Events:   adapter.Events{Q: q},
	}, p, checkFuncs)

	if err := svcs.Schedule.Start(ctx); err != nil {
		panic(err)
	}

	h := &Handlers{
		svcs:      svcs,
		users:     adapter.User{Q: q},
		access:    adapter.Access{Q: q},
		printer:   p,
		schedule:  svcs.Schedule,
	}

	mux := http.NewServeMux()
	protectedRoute := chainMiddleware(h.logRequestMiddleware, h.authMiddleware)

	mux.HandleFunc("POST /task", protectedRoute(h.createTask))
	mux.HandleFunc("PATCH /task/{id}", protectedRoute(h.doneTask))
	mux.HandleFunc("GET /task", protectedRoute(h.getOpenTasks))
	mux.HandleFunc("POST /list", protectedRoute(h.createList))
	mux.HandleFunc("POST /link", protectedRoute(h.createLink))
	mux.HandleFunc("GET /link/{id}", protectedRoute(h.getLink))
	mux.HandleFunc("PUT /end-of-day", protectedRoute(h.endOfDay))
	mux.HandleFunc("PATCH /end-of-day", protectedRoute(h.endOfDayAuto))
	mux.HandleFunc("POST /schedule", protectedRoute(h.createSchedule))
	mux.HandleFunc("GET /schedule", protectedRoute(h.getUserSchedule))
	mux.HandleFunc("PUT /schedule", protectedRoute(h.toggleSchedule))
	mux.HandleFunc("PUT /toggle-printer", protectedRoute(h.togglePrinter))
	mux.HandleFunc("POST /events", protectedRoute(h.postDayEvents))
	mux.HandleFunc("GET /health", h.healthCheck)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		slog.Info("shutting down")
		svcs.Schedule.Stop()
		os.Exit(1)
	}()

	server := &http.Server{
		Addr:    ":8000",
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// initLogger configures slog with a JSON handler at the level from LOG_LEVEL.
func initLogger() {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
