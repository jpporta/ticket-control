package printer

import (
	"fmt"
	"image"
	"log/slog"

	"github.com/jpporta/ticket-control/internal/printer/render"
)

type LinkInput struct {
	ID        int32
	Title     string
	URL       string
	CreatedBy string
}

func (p *Printer) PrintLink(link LinkInput) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintLink(link)
		})
		return fmt.Errorf("%w: queuing link: %s", errPrinterOffline, link.Title)
	}
	tmpl, ok := p.templates["link_header"]
	if !ok {
		return fmt.Errorf("link template not found")
	}

	pngFile, cleanup, err := render.Render(tmpl, link, "link")
	if err != nil {
		return err
	}
	defer cleanup()

	img, _, err := image.Decode(pngFile)
	pngFile.Close()
	if err != nil {
		return fmt.Errorf("decode link png: %w", err)
	}
	img = render.CropHeight8(img)

	close, err := p.start()
	if err != nil {
		slog.Error("printer start", "err", err)
		return err
	}
	defer close()
	p.Reset()
	if _, err := p.e.PrintImage(img); err != nil {
		return fmt.Errorf("print link image: %w", err)
	}
	p.e.WriteRaw([]byte{0x1b, 0x61, 0x01})
	if _, err := p.e.QRCode(link.URL, true, 10, 10); err != nil {
		return fmt.Errorf("print link qr: %w", err)
	}
	return p.e.PrintAndCut()
}
