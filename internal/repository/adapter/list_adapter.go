package adapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a List) CreateList(ctx context.Context, title, content string, createdBy int32) (int32, error) {
	return a.Q.CreateList(ctx, repository.CreateListParams{
		Title:     title,
		Content:   pgtype.Text{String: content, Valid: content != ""},
		CreatedBy: createdBy,
	})
}

func (a List) DeleteLastList(ctx context.Context, createdBy int32) error {
	return a.Q.DeleteLastList(ctx, createdBy)
}

func (a List) TotalListsFromUser(ctx context.Context, createdBy int32, start, end time.Time) (int64, error) {
	return a.Q.TotalListsFromUser(ctx, repository.TotalListsFromUserParams{
		CreatedBy:   createdBy,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
}
