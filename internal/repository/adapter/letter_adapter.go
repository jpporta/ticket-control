package adapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

func (a Letter) CreateLetter(ctx context.Context, title, content, recipient, sender string, createdBy int32) (int32, error) {
	return a.Q.CreateLetter(ctx, repository.CreateLetterParams{
		Title:     title,
		Content:   content,
		Recipient: recipient,
		Sender:    sender,
		CreatedBy: createdBy,
	})
}

func (a Letter) DeleteLastLetter(ctx context.Context, createdBy int32) error {
	return a.Q.DeleteLastLetter(ctx, createdBy)
}

func (a Letter) TotalLettersFromUser(ctx context.Context, createdBy int32, start, end time.Time) (int64, error) {
	return a.Q.TotalLettersFromUser(ctx, repository.TotalLettersFromUserParams{
		CreatedBy:   createdBy,
		CreatedAt:   pgtype.Timestamp{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: end, Valid: true},
	})
}
