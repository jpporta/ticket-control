package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

var LINK_LIMIT int64 = 50

func (a *Application) UserHasReachedLinkLimit(ctx context.Context, userId int32) (bool, error) {
	startYear, startMonth, startDay := time.Now().Date()
	startTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)

	total, err := a.Q.TotalLinksFromUser(ctx, repository.TotalLinksFromUserParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: startTime.Add(time.Hour * 24), Valid: true},
	})
	if err != nil {
		return false, err
	}

	return (total >= LINK_LIMIT), nil
}

func (a *Application) CreateLink(ctx context.Context, userId int32, title string, url string) (int32, error) {
	// Create in DB
	res, err := a.Q.CreateLink(ctx, repository.CreateLinkParams{
		Title:     title,
		Url:       url,
		CreatedBy: userId,
	})
	if err != nil {
		return 0, fmt.Errorf("Error creating link")
	}

	// Print, and if it fails, delete from DB
	name := ctx.Value("userName").(string)
	err = a.Printer.PrintLink(printer.LinkInput{
		Title:     title,
		URL:       url,
		CreatedBy: name})
	if err != nil {
		err_2 := a.Q.DeleteLastLink(ctx, userId)
		if err_2 != nil {
			return 0, fmt.Errorf("Error deleting link after printer start failure: %w", err)
		}
		return 0, fmt.Errorf("Error starting printer: %w", err)
	}

	return res, nil
}
