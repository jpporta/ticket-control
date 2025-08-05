package main

import (
	"log"
	"net/http"

	"github.com/jpporta/ticket-control/internal"
)

type Handlers struct {
	app *internal.Application
}

func (h *Handlers) createTask(w http.ResponseWriter, r *http.Request) {
	log.Println("Test", r.Context().Value("userId"))
}
