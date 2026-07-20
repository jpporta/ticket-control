package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/schedule"
)

func (h *Handlers) createSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	var params schedule.CreateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	if !h.schedule.IsValidCheckFunction(params.CheckFunction) {
		httperrWrite(w, r, fmt.Errorf("%w: invalid check function", apperr.ErrInvalidInput))
		return
	}
	if err := h.schedule.CreateSchedule(r.Context(), params, userID); err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handlers) getUserSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	rows, err := h.schedule.GetSchedules(r.Context(), userID)
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *Handlers) toggleSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	scheduleIDStr := r.URL.Query().Get("id")
	if scheduleIDStr == "" {
		httperrWrite(w, r, fmt.Errorf("%w: schedule id required", apperr.ErrInvalidInput))
		return
	}
	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	if err := h.schedule.ToggleSchedule(r.Context(), int32(scheduleID), userID); err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
