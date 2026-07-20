package adapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a Events) GetAccessStats(ctx context.Context, userID int32, path, method string, start, end time.Time) (int64, error) {
	return a.Q.GetAccessStats(ctx, repository.GetAccessStatsParams{
		UserID:       pgtype.Int4{Int32: userID, Valid: userID != 0},
		Path:         path,
		Method:       method,
		AccessedAt:   pgtype.Timestamp{Time: start, Valid: true},
		AccessedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
}

func (a User) CreateUser(ctx context.Context, name, apiKey string) (int32, error) {
	return a.Q.CreateUser(ctx, repository.CreateUserParams{Name: name, ApiKey: apiKey})
}

func (a User) GetUserByKey(ctx context.Context, key string) (int32, string, error) {
	row, err := a.Q.GetUserByKey(ctx, key)
	if err != nil {
		return 0, "", err
	}
	return row.ID, row.Name, nil
}

func (a Access) AddAccess(ctx context.Context, userID int32, ip, path, method string) error {
	return a.Q.AddAccess(ctx, repository.AddAccessParams{
		UserID:    pgtype.Int4{Int32: userID, Valid: userID != 0},
		IpAddress: ip,
		Path:      path,
		Method:    method,
	})
}
