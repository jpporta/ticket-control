package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/events"
)

// AppleLimitedBody is the nested-JSON envelope Apple Shortcuts sends.
type AppleLimitedBody struct {
	Body string `json:"body"`
}

type postDayEventsRequest struct {
	Events []events.Event `json:"events"`
}

func (h *Handlers) postDayEvents(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int32)
	var envelope AppleLimitedBody
	if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	var payload postDayEventsRequest
	if err := json.Unmarshal([]byte(envelope.Body), &payload); err != nil {
		httperrWrite(w, r, fmt.Errorf("%w: %v", apperr.ErrInvalidInput, err))
		return
	}
	noCreated, err := h.svcs.Events.CreateEvents(r.Context(), payload.Events, userID)
	if err != nil {
		httperrWrite(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"created": noCreated})
}
