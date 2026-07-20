package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/clock"
)

type middleware func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func (h *Handlers) logRequestMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ip := r.RemoteAddr
	key := r.Header.Get("x-api-key")
	start := clock.Now()
	defer func() {
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", ip,
			"took", time.Since(start).String(),
		)
	}()
	if key == "" {
		httperrWrite(w, r, apperr.ErrUnauthorized)
		return
	}
	userID, userName, err := h.users.GetUserByKey(r.Context(), key)
	if err != nil {
		slog.Warn("get user by key", "err", err)
	}
	if err := h.access.AddAccess(r.Context(), userID, ip, r.RequestURI, r.Method); err != nil {
		slog.Warn("add access", "err", err)
	}
	ctx := context.WithValue(
		context.WithValue(r.Context(), ctxKeyUserID, userID),
		ctxKeyUserName, userName)
	r = r.WithContext(ctx)
	next(w, r)
}

func (h *Handlers) authMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	id := r.Context().Value(ctxKeyUserID).(int32)
	if id == 0 {
		httperrWrite(w, r, apperr.ErrUnauthorized)
		return
	}
	next(w, r)
}
