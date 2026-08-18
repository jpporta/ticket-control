package printer

import (
	"bytes"
	"fmt"
	"image"
	"log/slog"
	"text/template"

	"github.com/jpporta/ticket-control/internal/printer/render"
	"github.com/jpporta/ticket-control/internal/utils"
)

type LetterInput struct {
	Title    string
	Date     string
	To       string
	ToLabel  string
	From     string
	SignOff  string
	Font     string
	FontSize string
	Justify  bool
	Content  string
}

// RenderLetter processes markdown content and renders the letter template to a
// cropped ESC/POS-ready image.Image without requiring a printer or database connection.
func RenderLetter(input LetterInput) (image.Image, error) {
	raw, err := models.ReadFile("models/letter.typ")
	if err != nil {
		return nil, fmt.Errorf("read letter template: %w", err)
	}

	tmpl, err := template.New("letter").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse letter template: %w", err)
	}

	data := input
	if data.ToLabel == "" && data.To != "" {
		data.ToLabel = "Para"
	}
	data.Content = utils.MarkdownToTypst(input.Content)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute letter template: %w", err)
	}

	return render.RenderSourceToImage(buf.String(), "letter")
}

func (p *Printer) PrintLetter(input LetterInput) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintLetter(input)
		})
		return fmt.Errorf("%w: queuing letter", errPrinterOffline)
	}

	img, err := RenderLetter(input)
	if err != nil {
		return err
	}

	close, err := p.start()
	if err != nil {
		slog.Error("printer start", "err", err)
		return err
	}
	defer close()
	p.Reset()
	return p.printImage(img)
}
