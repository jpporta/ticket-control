package printer

import (
	"fmt"
	"image"
)

func (p *Printer) printImage(img image.Image) error {
	_, err := p.e.PrintImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return p.e.PrintAndCut()
}
