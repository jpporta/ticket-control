package printer

import (
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/jpporta/ticket-control/internal/printer/render"
)

type EndOfDayInput struct {
	CreatedBy string
	Day       time.Time
	EndDay    time.Time
	NoTasks   int
	NoDone    int
}

func (p *Printer) PrintEndOfDay(input EndOfDayInput) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintEndOfDay(input)
		})
		return fmt.Errorf("%w: queuing end of day", errPrinterOffline)
	}
	tmpl, ok := p.templates["end_of_day"]
	if !ok {
		return fmt.Errorf("end_of_day template not found")
	}

	pngFile, cleanup, err := render.Render(tmpl, input, "eod")
	if err != nil {
		return err
	}
	defer cleanup()

	img, _, err := image.Decode(pngFile)
	pngFile.Close()
	if err != nil {
		return fmt.Errorf("decode eod png: %w", err)
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
