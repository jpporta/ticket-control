package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jpporta/ticket-control/internal/apperr"
)

func (h *Handlers) endOfDay(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	userName := r.Context().Value(ctxKeyUserName).(string)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: read body: %v", apperr.ErrInvalidInput, err))
		return
	}
	offset, err := atoiOrZero(r.URL.Query().Get("offset"))
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: offset: %v", apperr.ErrInvalidInput, err))
		return
	}
	noDone, err := strconv.Atoi(string(body))
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: noDone: %v", apperr.ErrInvalidInput, err))
		return
	}
	if err := h.svcs.EndOfDay.EndOfDay(r.Context(), userID, userName, noDone, offset); err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "end of day processed"})
}

func (h *Handlers) endOfDayAuto(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	userName := r.Context().Value(ctxKeyUserName).(string)
	offset, err := atoiOrZero(r.URL.Query().Get("offset"))
	if err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: offset: %v", apperr.ErrInvalidInput, err))
		return
	}
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr != "" && endStr != "" {
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			httperrWrite(w, r, fmt.Errorf("%w: start: %v", apperr.ErrInvalidInput, err))
			return
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			httperrWrite(w, r, fmt.Errorf("%w: end: %v", apperr.ErrInvalidInput, err))
			return
		}
		if err := h.svcs.EndOfDay.EndOfWeekend(r.Context(), userID, userName, start, end); err != nil {
			httperrWrite(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "end of day processed"})
		return
	}
	if err := h.svcs.EndOfDay.EndOfDayAuto(r.Context(), userID, userName, offset); err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "end of day processed"})
}

func atoiOrZero(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
