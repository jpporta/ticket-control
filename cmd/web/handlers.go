package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jpporta/ticket-control/internal"
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

	limitReached, err := h.app.UserHasReachedTaskLimit(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error checking task limit", http.StatusInternalServerError)
		return
	}

	if limitReached {
		http.Error(w, "You have reached your task limit for today", http.StatusForbidden)
		return
	}

	task := CreateTask{
		Priority: 1,
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&task)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	total, err := h.app.CreateTask(r.Context(), task.Title, task.Description, task.Priority, userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating task: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": %d}`, total)
}
