package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Handlers struct {
	app *internal.Application
}

type CreateTask struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    int32  `json:"priority,omitempty"`
}

func (h *Handlers) createTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)
	task := CreateTask{
		Priority: 1,
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil || task.Title == "" {
		http.Error(w, "Task does not look right", http.StatusBadRequest)
	}
	res, err := h.app.Q.CreateTask(r.Context(), repository.CreateTaskParams{
		Title:       task.Title,
		Description: pgtype.Text{String: task.Description, Valid: task.Description != ""},
		Priority:    pgtype.Int4{Int32: task.Priority, Valid: task.Priority > 0 && task.Priority <= 5},
		CreatedBy: userId,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%v", res)
}
