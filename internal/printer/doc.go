package printer

import (
	"fmt"
	"image"
	"log/slog"

	"github.com/jpporta/ticket-control/internal/printer/render"
)

// PrintImage prints an image and cuts the paper.
func (p *Printer) PrintImage(img image.Image) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintImage(img)
		})
		return fmt.Errorf("%w: queuing image", ErrPrinterOffline)
	}

	img = PrepareImage(img)
	close, err := p.start()
	if err != nil {
		slog.Error("printer start", "err", err)
		return err
	}
	defer close()

	p.Reset()
	return p.printImage(img)
}

// PrintTypst compiles Typst source code and prints the resulting raster image.
func (p *Printer) PrintTypst(source string) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintTypst(source)
		})
		return fmt.Errorf("%w: queuing typst doc", ErrPrinterOffline)
	}

	img, err := render.RenderSourceToImage(source, "doc")
	if err != nil {
		return err
	}

	return p.PrintImage(img)
}

// PrintTypstFile compiles an existing .typ file and prints the resulting raster image.
func (p *Printer) PrintTypstFile(filePath string) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintTypstFile(filePath)
		})
		return fmt.Errorf("%w: queuing typst file", ErrPrinterOffline)
	}

	img, err := render.RenderFileToImage(filePath)
	if err != nil {
		return err
	}

	return p.PrintImage(img)
}
