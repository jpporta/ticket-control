package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpporta/ticket-control/internal"
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
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	app := internal.New(conn)
	h := Handlers{
		app,
	}

	mux := http.NewServeMux()

	protectedRoute := chainMiddleware(h.logRequestMiddleware, h.authMiddleware)

	mux.HandleFunc("POST /task", protectedRoute(h.createTask))
	mux.HandleFunc("POST /list", protectedRoute(h.createList))
	mux.HandleFunc("POST /link", protectedRoute(h.createLink))
	mux.HandleFunc("PUT /end-of-day", protectedRoute(h.endOfDay))
	mux.HandleFunc("POST /schedule", protectedRoute(h.createSchedule))
	mux.HandleFunc("GET /schedule", protectedRoute(h.getUserSchedule))
	mux.HandleFunc("PUT /schedule", protectedRoute(h.toggleSchedule))
	mux.HandleFunc("GET /health", h.healthCheck)

	err = app.Cron.Start(ctx, app)
	if err != nil {
		panic(err)
	}

	signal := make(chan os.Signal, 1)
	go func() {
		sig := <-signal
		log.Println("Received signal:", sig)
		app.Cron.Stop()
		if err != nil {
			log.Println("Error stopping cron job:", err)
		}
		os.Exit(1)
	}()

	err = http.ListenAndServe(":8000", mux)
	if err != nil {
		panic(err)
	}
}
