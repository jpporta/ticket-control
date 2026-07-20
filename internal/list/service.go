package list

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/ports"
	"github.com/jpporta/ticket-control/internal/printer"
)

type Service struct {
	repo    Repo
	printer ports.Printer
}

func New(repo Repo, p ports.Printer) *Service {
	return &Service{repo: repo, printer: p}
}

func (s *Service) UserHasReachedListLimit(ctx context.Context, userID int32) (bool, error) {
	start := clock.Today()
	total, err := s.repo.TotalListsFromUser(ctx, userID, start, start.Add(24*60*60*1e9))
	if err != nil {
		return false, fmt.Errorf("count user lists: %w", err)
	}
	return total >= ListLimit, nil
}

func (s *Service) CreateList(ctx context.Context, params CreateParams, userName string, userID int32) (int32, error) {
	limitReached, err := s.UserHasReachedListLimit(ctx, userID)
	if err != nil {
		return 0, err
	}
	if limitReached {
		return 0, fmt.Errorf("list: %w", apperr.ErrQuotaExceeded)
	}

	id, err := s.repo.CreateList(ctx, params.Title, strings.Join(params.Items, "\n"), userID)
	if err != nil {
		return 0, fmt.Errorf("create list: %w", err)
	}

	if err := s.printer.PrintList(printer.ListInput{
		Title:     params.Title,
		Content:   params.Items,
		CreatedBy: userName,
	}); err != nil {
		_ = s.repo.DeleteLastList(ctx, userID)
		if errors.Is(err, errPrinterOffline) {
			return 0, fmt.Errorf("print list: %w", apperr.ErrPrinterOffline)
		}
		return 0, fmt.Errorf("print list: %w", err)
	}
	return id, nil
}

var errPrinterOffline = errors.New("printer offline")
