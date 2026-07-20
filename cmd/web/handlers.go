package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/link"
	"github.com/jpporta/ticket-control/internal/list"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository/adapter"
	"github.com/jpporta/ticket-control/internal/schedule"
	"github.com/jpporta/ticket-control/internal/service"
	"github.com/jpporta/ticket-control/internal/task"
)

type Handlers struct {
	svcs     *service.Services
	users    adapter.User
	access   adapter.Access
	printer  *printer.Printer
	schedule *schedule.Service
}

type ctxKey string

const (
	ctxKeyUserID   ctxKey = "userId"
	ctxKeyUserName ctxKey = "userName"
)

// --- task ---

type createTaskReq struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    int32  `json:"priority,omitempty"`
}

func (h *Handlers) createTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	var req createTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	if req.Priority == 0 {
		req.Priority = 1
	}
	id, err := h.svcs.Task.CreateTask(r.Context(), task.CreateParams{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	}, userID)
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int32{"id": id})
}

func (h *Handlers) doneTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	if err := h.svcs.Task.MarkTaskAsDone(r.Context(), int32(taskID), userID); err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) getOpenTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}
	tasks, err := h.svcs.Task.GetOpenTasks(r.Context(), userID, limit, page)
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// --- list ---

type createListReq struct {
	Title string `json:"title"`
	Items string `json:"items"`
}

func listItems(items string) []string {
	var result []string
	for _, item := range strings.Split(items, "\n") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func (h *Handlers) createList(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	userName := r.Context().Value(ctxKeyUserName).(string)
	var req createListReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	items := listItems(req.Items)
	if len(items) == 0 {
		httperrWrite(w, r, apperr.ErrInvalidInput)
		return
	}
	id, err := h.svcs.List.CreateList(r.Context(), list.CreateParams{
		Title: req.Title,
		Items: items,
	}, userName, userID)
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int32{"id": id})
}

// --- link ---

type createLinkReq struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func (h *Handlers) createLink(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	userName := r.Context().Value(ctxKeyUserName).(string)
	var req createLinkReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	id, err := h.svcs.Link.CreateLink(r.Context(), link.CreateParams{
		Title: req.Title,
		URL:   req.URL,
	}, userName, userID)
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int32{"id": id})
}

func (h *Handlers) getLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	url, err := h.svcs.Link.GetLink(r.Context(), int32(id))
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// --- printer ---

func (h *Handlers) togglePrinter(w http.ResponseWriter, r *http.Request) {
	h.printer.Toggle(!h.printer.Enabled)
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": h.printer.Enabled})
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"now":    clock.Now().Format(time.RFC3339),
	})
}
