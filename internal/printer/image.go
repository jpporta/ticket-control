package printer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

const maxImageWidth = 576

// PrepareImage makes an uploaded image fit an 80mm printer and aligns both
// dimensions to the ESC/POS raster format. The escpos library thresholds the
// resulting image to black and white when it sends it.
func PrepareImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width > maxImageWidth {
		height = height * maxImageWidth / width
		width = maxImageWidth
	} else {
		width = (width + 7) / 8 * 8
	}
	height = height / 8 * 8
	if height < 8 {
		height = 8
	}

	prepared := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(prepared, prepared.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	for y := 0; y < height; y++ {
		for x := 0; x < width && x < bounds.Dx(); x++ {
			sx := bounds.Min.X + x*bounds.Dx()/width
			sy := bounds.Min.Y + y*bounds.Dy()/height
			prepared.Set(x, y, img.At(sx, sy))
		}
	}
	return prepared
}

func (p *Printer) printImage(img image.Image) error {
	_, err := p.e.PrintImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return p.e.PrintAndCut()
}
