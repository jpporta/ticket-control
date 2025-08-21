package main

import (
	"slices"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jpporta/ticket-control/internal"
)

func (h *Handlers) createSchedule(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)

	var schedule internal.Schedule
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&schedule)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	schedule.UserId = userId

	// Validate CheckFunction
	validCheckFunction := true
	if schedule.CheckFunction != "" {
		validCheckFunction = slices.Contains(internal.PossibleCheckFunctions, schedule.CheckFunction)
	}

	if !validCheckFunction {
		http.Error(w, "Invalid check function", http.StatusBadRequest)
		return
	}

	err = h.app.CreateSchedule(r.Context(), &schedule)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating schedule: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "OK")
}

func (h *Handlers) getUserSchedule(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)

	schedules, err := h.app.GetSchedules(r.Context(), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching schedules: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedules); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) toggleSchedule(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int32)

	scheduleIdStr := r.URL.Query().Get("id")
	if scheduleIdStr == "" {
		http.Error(w, "Schedule ID is required", http.StatusBadRequest)
		return
	}

	scheduleId, err := strconv.Atoi(scheduleIdStr)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	err = h.app.ToggleSchedule(r.Context(), int32(scheduleId), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error toggling schedule: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
