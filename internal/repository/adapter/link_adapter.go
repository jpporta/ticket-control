package adapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a Link) CreateLink(ctx context.Context, title, url string, createdBy int32) (int32, error) {
	return a.Q.CreateLink(ctx, repository.CreateLinkParams{
		Url:       url,
		Title:     title,
		CreatedBy: createdBy,
	})
}

func (a Link) DeleteLastLink(ctx context.Context, createdBy int32) error {
	return a.Q.DeleteLastLink(ctx, createdBy)
}

func (a Link) GetLinkByID(ctx context.Context, id int32) (string, error) {
	return a.Q.GetLinkByID(ctx, id)
}

func (a Link) TotalLinksFromUser(ctx context.Context, createdBy int32, start, end time.Time) (int64, error) {
	return a.Q.TotalLinksFromUser(ctx, repository.TotalLinksFromUserParams{
		CreatedBy:   createdBy,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
}
