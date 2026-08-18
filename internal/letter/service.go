package letter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jpporta/ticket-control/internal/apperr"
	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/ports"
	"github.com/jpporta/ticket-control/internal/printer"
)

var monthsPT = [...]string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

func formatDatePT(t time.Time) string {
	return fmt.Sprintf("%d de %s de %d", t.Day(), monthsPT[t.Month()-1], t.Year())
}

type Service struct {
	repo    Repo
	printer ports.Printer
}

func New(repo Repo, p ports.Printer) *Service {
	return &Service{repo: repo, printer: p}
}

func (s *Service) UserHasReachedLetterLimit(ctx context.Context, userID int32) (bool, error) {
	start := clock.Today()
	total, err := s.repo.TotalLettersFromUser(ctx, userID, start, start.Add(24*60*60*1e9))
	if err != nil {
		return false, fmt.Errorf("count user letters: %w", err)
	}
	return total >= LetterLimit, nil
}

func (s *Service) CreateLetter(ctx context.Context, params CreateParams, userName string, userID int32) (int32, error) {
	limitReached, err := s.UserHasReachedLetterLimit(ctx, userID)
	if err != nil {
		return 0, err
	}
	if limitReached {
		return 0, fmt.Errorf("letter: %w", apperr.ErrQuotaExceeded)
	}

	from := params.From
	if from == "" {
		from = userName
	}

	id, err := s.repo.CreateLetter(ctx, params.Title, params.Content, params.To, from, userID)
	if err != nil {
		return 0, fmt.Errorf("create letter: %w", err)
	}

	dateVal := params.Date
	if dateVal == "" {
		dateVal = formatDatePT(clock.Now())
	}

	toLabel := params.ToLabel
	if toLabel == "" && params.To != "" {
		toLabel = "Para"
	}

	signOff := params.SignOff
	if signOff == "" && from != "" {
		signOff = "Atenciosamente,"
	}

	font := params.Font
	if font == "" {
		font = "Libertinus Serif"
	}

	fontSize := params.FontSize
	if fontSize == "" {
		fontSize = "11pt"
	}

	if err := s.printer.PrintLetter(printer.LetterInput{
		Title:    params.Title,
		Date:     dateVal,
		To:       params.To,
		ToLabel:  toLabel,
		From:     from,
		SignOff:  signOff,
		Font:     font,
		FontSize: fontSize,
		Justify:  params.Justify,
		Content:  params.Content,
	}); err != nil {
		_ = s.repo.DeleteLastLetter(ctx, userID)
		if errors.Is(err, errPrinterOffline) {
			return 0, fmt.Errorf("print letter: %w", apperr.ErrPrinterOffline)
		}
		return 0, fmt.Errorf("print letter: %w", err)
	}
	return id, nil
}

var errPrinterOffline = errors.New("printer offline")
