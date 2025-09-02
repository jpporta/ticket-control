package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

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
	err = h.app.EndOfDayWithTasks(r.Context(), userId, userName, req.DoneTasks)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error ending day: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}
