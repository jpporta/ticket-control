package printer

import (
	"image"
	"image/color"
	"testing"
)

func TestPrepareImageFitsPrinterAndAlignsRaster(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1200, 1000))
	img.Set(0, 0, color.Black)
	got := PrepareImage(img).Bounds()
	if got.Dx() != 576 || got.Dy() != 480 {
		t.Fatalf("prepared image bounds = %v, want 576x480", got)
	}
}
