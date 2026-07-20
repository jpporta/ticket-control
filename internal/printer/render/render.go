package render

import (
	"fmt"
	"image"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// Render executes the given template with data, shells out to typst to compile
// it to a PNG, decodes the PNG, and returns a ReadCloser on the image file
// plus a cleanup func that removes both the .typ and the .png from disk.
//
// Callers MUST invoke the returned cleanup func.
//
// The "name" parameter is used in the temp filename prefix for easier debugging
// (e.g. "task", "list").
func Render(tmpl *template.Template, data any, name string) (io.ReadCloser, func(), error) {
	typFile, err := os.CreateTemp("", name+"-*.typ")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp %s typ: %w", name, err)
	}
	typPath := typFile.Name()
	pngPath := strings.TrimSuffix(typPath, ".typ") + ".png"

	cleanup := func() {
		os.Remove(typPath)
		os.Remove(pngPath)
	}

	if err := tmpl.Execute(typFile, data); err != nil {
		typFile.Close()
		cleanup()
		return nil, nil, fmt.Errorf("execute %s template: %w", name, err)
	}
	if err := typFile.Close(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("close %s typ: %w", name, err)
	}

	cmd := exec.Command("typst", "c", typPath, "-f", "png")
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("typst compile %s: %w (%s)", name, err, string(out))
	}

	pngFile, err := os.Open(pngPath)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("open %s png: %w", name, err)
	}
	return pngFile, cleanup, nil
}

// CropHeight8 trims the bottom of the image so its height is a multiple of 8
// pixels, as required by the ESC/POS raster command.
func CropHeight8(img image.Image) image.Image {
	b := img.Bounds()
	if b.Max.Y%8 == 0 {
		return img
	}
	cropRect := image.Rect(0, 0, b.Max.X, b.Max.Y-(b.Max.Y%8))
	if si, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		return si.SubImage(cropRect)
	}
	return img
}
