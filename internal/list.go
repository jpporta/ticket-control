package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

var LIST_LIMIT int64 = 10

func (a *Application) UserHasReachedListLimit(ctx context.Context, userId int32) (bool, error) {
	startYear, startMonth, startDay := time.Now().Date()
	startTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)

	total, err := a.Q.TotalListsFromUser(ctx, repository.TotalListsFromUserParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: startTime.Add(time.Hour * 24), Valid: true},
	})
	if err != nil {
		return false, err
	}

	return (total >= LIST_LIMIT), nil
}

func (a *Application) CreateList(ctx context.Context, userId int32, title string, items []string) (int32, error) {
	content := ""
	for _, item := range items {
		content += item + "\n"
	}
	// Create in DB
	res, err := a.Q.CreateList(ctx, repository.CreateListParams{
		Title:     title,
		Content:   pgtype.Text{String: content, Valid: content != ""},
		CreatedBy: userId,
	})
	if err != nil {
		return 0, fmt.Errorf("Error creating list")
	}

	// Print, and if it fails, delete from DB
	p, err := printerInternal.New()
	if err != nil {
		err_2 := a.Q.DeleteLastList(ctx, userId)
		if err_2 != nil {
			return 0, fmt.Errorf("Error deleting list after printer start failure: %w", err)
		}
		return 0, fmt.Errorf("Error starting printer: %w", err)
	}
	name := ctx.Value("userName").(string)
	err = p.PrintList(printerInternal.ListInput{
		Title:     title,
		Content:   items,
		CreatedBy: name,
	})
	return res, nil
}
