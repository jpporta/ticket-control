package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jpporta/ticket-control/internal"
)

type AppleLimitedBody struct {
	Body string `json:"body"`
}

type postDayEventsRequest struct {
	Events []internal.Event `json:"events"`
}

func (h *Handlers) postDayEvents(w http.ResponseWriter, r *http.Request) {
	var body AppleLimitedBody
	json.NewDecoder(r.Body).Decode(&body)
	var payload postDayEventsRequest
	json.Unmarshal([]byte(body.Body), &payload)
	noCreated, err := h.app.CreateEvents(r.Context(), payload.Events, r.Context().Value("userId").(int32))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating events: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%d", noCreated)
}
