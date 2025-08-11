package printer

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png"
)

type Server struct{}

func (s *Server) Print(ctx context.Context, in *PrintJob) (*Empty, error) {
	p := New(ctx)
	if p == nil {
		return nil, fmt.Errorf("printer not initialized")
	}
	if !p.Enabled {
		return nil, fmt.Errorf("printer is not enabled")
	}
	close, err := p.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start printer: %w", err)
	}
	defer close()

	img, _, err := image.Decode(bytes.NewReader(in.Img))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	if err := p.PrintImage(img); err != nil {
		return nil, fmt.Errorf("failed to print image: %w", err)
	}
	return nil, nil
}

func (s *Server) PrintLink(ctx context.Context, link *PrintLinkJob) (*Empty, error) {
	p := New(ctx)
	if p == nil {
		return nil, fmt.Errorf("printer not initialized")
	}
	if !p.Enabled {
		return nil, fmt.Errorf("printer is not enabled")
	}
	close, err := p.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start printer: %w", err)
	}
	defer close()

	img, _, err := image.Decode(bytes.NewReader(link.Header))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	p.e.PrintImage(img)
	p.e.WriteRaw([]byte{0x1b, 0x61, 0x01})
	_, err = p.e.QRCode(link.Url, true, 10, 10)
	if err != nil {
		return nil, fmt.Errorf("error printing qr: %w", err)
	}
	return nil, p.e.PrintAndCut()
}
