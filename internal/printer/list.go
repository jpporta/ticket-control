package printer

import (
	"fmt"
	"image"
	"log/slog"

	"github.com/jpporta/ticket-control/internal/printer/render"
)

type ListInput struct {
	Title     string
	Content   []string
	CreatedBy string
}

func (p *Printer) PrintList(list ListInput) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintList(list)
		})
		return fmt.Errorf("%w: queuing list: %s", errPrinterOffline, list.Title)
	}
	tmpl, ok := p.templates["list"]
	if !ok {
		return fmt.Errorf("list template not found")
	}

	pngFile, cleanup, err := render.Render(tmpl, list, "list")
	if err != nil {
		return err
	}
	defer cleanup()

	img, _, err := image.Decode(pngFile)
	pngFile.Close()
	if err != nil {
		return fmt.Errorf("decode list png: %w", err)
	}
	img = render.CropHeight8(img)

	close, err := p.start()
	if err != nil {
		slog.Error("printer start", "err", err)
		return err
	}
	defer close()
	p.Reset()
	return p.printImage(img)
}
