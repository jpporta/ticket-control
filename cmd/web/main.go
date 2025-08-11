package main

import (
	"context"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
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
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())
	app := internal.New(conn)
	h := Handlers{
		app,
	}

	mux := http.NewServeMux()

	protectedRoute := chainMiddleware(h.logRequestMiddleware, h.authMiddleware)

	mux.HandleFunc("POST /task", protectedRoute(h.createTask))
	mux.HandleFunc("POST /list", protectedRoute(h.createList))
	mux.HandleFunc("POST /link", protectedRoute(h.createLink))
	mux.HandleFunc("GET /health", h.healthCheck)

	err = http.ListenAndServe(":8000", mux)
	if err != nil {
		panic(err)
	}
}
