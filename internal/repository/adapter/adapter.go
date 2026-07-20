// Package adapter exposes the sqlc-generated *repository.Queries through
// per-domain interfaces. Each domain service depends only on its own Repo
// interface; this package is the only place that converts pgtype wrappers
// to plain Go types for the domain boundary.
package adapter

import (
	"github.com/jpporta/ticket-control/internal/repository"
)

// Per-domain adapter types. The methods on each live in <domain>_adapter.go
// alongside the sqlc query they wrap.
type (
	Task     struct{ Q *repository.Queries }
	List     struct{ Q *repository.Queries }
	Link     struct{ Q *repository.Queries }
	Schedule struct{ Q *repository.Queries }
	EndOfDay struct{ Q *repository.Queries }
	Events   struct{ Q *repository.Queries }
	User     struct{ Q *repository.Queries }
	Access   struct{ Q *repository.Queries }
)
