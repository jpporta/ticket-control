package main

import (
	"log/slog"
	"net/http"

	"github.com/jpporta/ticket-control/internal/apperr"
)

// httperrWrite maps a wrapped error to an HTTP status code and writes a fixed
// JSON body. It never leaks internal error text to the client.
func httperrWrite(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		return
	case apperr.Kind(err) == apperr.ErrQuotaExceeded:
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "quota_exceeded"})
	case apperr.Kind(err) == apperr.ErrPrinterOffline:
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "printer_offline"})
	case apperr.Kind(err) == apperr.ErrNotFound:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
	case apperr.Kind(err) == apperr.ErrUnauthorized:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	case apperr.Kind(err) == apperr.ErrInvalidInput:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_input"})
	default:
		slog.Error("handler error",
			"path", r.URL.Path,
			"method", r.Method,
			"err", err,
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = jsonEncode(w, body)
}
