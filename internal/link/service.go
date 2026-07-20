package link

import (
	"context"
	"errors"
	"fmt"

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

func (s *Service) UserHasReachedLinkLimit(ctx context.Context, userID int32) (bool, error) {
	start := clock.Today()
	total, err := s.repo.TotalLinksFromUser(ctx, userID, start, start.Add(24*60*60*1e9))
	if err != nil {
		return false, fmt.Errorf("count user links: %w", err)
	}
	return total >= LinkLimit, nil
}

func (s *Service) CreateLink(ctx context.Context, params CreateParams, userName string, userID int32) (int32, error) {
	limitReached, err := s.UserHasReachedLinkLimit(ctx, userID)
	if err != nil {
		return 0, err
	}
	if limitReached {
		return 0, fmt.Errorf("link: %w", apperr.ErrQuotaExceeded)
	}

	id, err := s.repo.CreateLink(ctx, params.Title, params.URL, userID)
	if err != nil {
		return 0, fmt.Errorf("create link: %w", err)
	}

	if err := s.printer.PrintLink(printer.LinkInput{
		ID:        id,
		Title:     params.Title,
		URL:       params.URL,
		CreatedBy: userName,
	}); err != nil {
		_ = s.repo.DeleteLastLink(ctx, userID)
		if errors.Is(err, errPrinterOffline) {
			return 0, fmt.Errorf("print link: %w", apperr.ErrPrinterOffline)
		}
		return 0, fmt.Errorf("print link: %w", err)
	}
	return id, nil
}

func (s *Service) GetLink(ctx context.Context, linkID int32) (string, error) {
	url, err := s.repo.GetLinkByID(ctx, linkID)
	if err != nil {
		return "", fmt.Errorf("get link: %w", err)
	}
	if url == "" {
		return "", fmt.Errorf("link %d: %w", linkID, apperr.ErrNotFound)
	}
	return url, nil
}

var errPrinterOffline = errors.New("printer offline")
