package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

type middleware func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func (h *Handlers) logRequestMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ip := r.RemoteAddr
	key := r.Header.Get("x-api-key")

	start := time.Now()

	defer func() {
		log.Println("[", r.Method, " @ ", r.Proto, "] - ", r.RequestURI, " - ", r.RemoteAddr, " - Took:", time.Since(start))
	}()
	if key == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Missing API key"))
		return
	}
	user, err := h.app.Q.GetUserByKey(r.Context(), key)
	if err != nil {
		log.Println("Error getting user by key:", err)
	}
	h.app.Q.AddAccess(r.Context(), repository.AddAccessParams{
		UserID:    pgtype.Int4{Int32: user.ID, Valid: user.ID != 0},
		IpAddress: ip,
		Path:      r.RequestURI,
		Method:    r.Method,
	})
	ctx := context.WithValue(
		context.WithValue(
			r.Context(), "userId", user.ID,
		),
		"userName", user.Name)
	r = r.WithContext(ctx)
	next(w, r)
}

func (h *Handlers) authMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	id := r.Context().Value("userId").(int32)
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	next(w, r)
}
