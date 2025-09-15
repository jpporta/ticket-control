package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (h *Handlers) endOfWeekend(w http.ResponseWriter, r *http.Request) {
	userName := r.Context().Value("userName").(string)
	userId := r.Context().Value("userId").(int32)
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading start: %v", err), http.StatusBadRequest)
		return
	}
	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading end: %v", err), http.StatusBadRequest)
		return
	}
	err = h.app.EndOfWeekend(r.Context(), userId, userName, start, end)
}

func (h *Handlers) endOfDay(w http.ResponseWriter, r *http.Request) {
	userName := r.Context().Value("userName").(string)
	userId := r.Context().Value("userId").(int32)
	done, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading offset: %v", err), http.StatusBadRequest)
			return
		}
	}

	noDone, err := strconv.Atoi(string(done))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}

	err = h.app.EndOfDay(r.Context(), userId, userName, noDone, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error ending day: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\": \"end of day processed\"}"))
}

type endOfDayWithTasksRequest struct {
	DoneTasks []int32 `json:"done"`
}

func (h *Handlers) endOfDayWithTasks(w http.ResponseWriter, r *http.Request) {
	userName := r.Context().Value("userName").(string)
	userId := r.Context().Value("userId").(int32)
	var req endOfDayWithTasksRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}
	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading offset: %v", err), http.StatusBadRequest)
			return
		}
	}
	err = h.app.EndOfDayWithTasks(r.Context(), userId, userName, req.DoneTasks, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error ending day: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func (h *Handlers) endOfDayAuto(w http.ResponseWriter, r *http.Request) {
	userName := r.Context().Value("userName").(string)
	userId := r.Context().Value("userId").(int32)
	offsetStr := r.URL.Query().Get("offset")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	offset := 0
	err := error(nil)
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading offset: %v", err), http.StatusBadRequest)
			return
		}
	}

	var start, end time.Time
	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading start: %v", err), http.StatusBadRequest)
		return
	}
	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading end: %v", err), http.StatusBadRequest)
		return
	}

	if start.IsZero() {
		err = h.app.EndOfDayAuto(r.Context(), userId, userName, offset)

	} else {
		if end.IsZero() {
			now := time.Now()
			end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		}
		err = h.app.EndOfWeekend(r.Context(), userId, userName, start, end)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error ending day: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\": \"end of day processed\"}"))
}
