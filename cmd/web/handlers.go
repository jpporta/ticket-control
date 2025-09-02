package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jpporta/ticket-control/internal"
)

type Handlers struct {
	app *internal.Application
}

type CreateLink struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (h *Handlers) getLink(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid link ID", http.StatusBadRequest)
		return
	}
	link, err := h.app.GetLink(r.Context(), int32(id), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving link: %v", err), http.StatusInternalServerError)
		return
	}
	if link == "" {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, link)
}

func (h *Handlers) createLink(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)

	limitReached, err := h.app.UserHasReachedLinkLimit(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error checking link limit", http.StatusInternalServerError)
		return
	}
	if limitReached {
		http.Error(w, "You have reached your link limit for today", http.StatusForbidden)
		return
	}

	var link CreateLink
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&link)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	id, err := h.app.CreateLink(r.Context(), userId, link.Title, link.Url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating link: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": %d}`, id)
}

type CreateList struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

func (h *Handlers) createList(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)

	limitReached, err := h.app.UserHasReachedListLimit(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error checking list limit", http.StatusInternalServerError)
		return
	}
	if limitReached {
		http.Error(w, "You have reached your list limit for today", http.StatusForbidden)
		return
	}

	var list CreateList
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&list)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	id, err := h.app.CreateList(r.Context(), userId, list.Title, list.Items)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating list: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": %d}`, id)
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

func (h *Handlers) getOpenTasks(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)
	tasks, err := h.app.GetOpenTasks(r.Context(), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving tasks: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "ok", "now": time.Now().Local().Format(time.RFC3339)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) endOfDay(w http.ResponseWriter, r *http.Request) {
	userName := r.Context().Value("userName").(string)
	userId := r.Context().Value("userId").(int32)
	done, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}

	noDone, err := strconv.Atoi(string(done))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}

	err = h.app.EndOfDay(r.Context(), userId, userName, noDone)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error ending day: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\": \"end of day processed\"}"))
}

func (h *Handlers) togglePrinter(w http.ResponseWriter, _ *http.Request) {
	h.app.Printer.TooglePrinter(!h.app.Printer.Enabled)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]bool{"enabled": h.app.Printer.Enabled}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}
