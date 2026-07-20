package apperr

import "errors"

var (
	ErrQuotaExceeded   = errors.New("quota exceeded")
	ErrPrinterOffline  = errors.New("printer offline")
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidInput    = errors.New("invalid input")
)

// Kind walks the error chain with errors.Is and returns the first matching
// sentinel, or nil if none match. Use this from the HTTP layer to map errors
// to status codes.
func Kind(err error) error {
	for _, target := range []error{ErrQuotaExceeded, ErrPrinterOffline, ErrNotFound, ErrUnauthorized, ErrInvalidInput} {
		if errors.Is(err, target) {
			return target
		}
	}
	return nil
}
