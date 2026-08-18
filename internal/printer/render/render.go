package render

import (
	"bytes"
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
// it to a PNG, and returns a ReadCloser on the image file plus a cleanup func
// that removes temporary files.
//
// Callers MUST invoke the returned cleanup func.
func Render(tmpl *template.Template, data any, name string) (io.ReadCloser, func(), error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, nil, fmt.Errorf("execute %s template: %w", name, err)
	}
	return RenderSource(buf.String(), name)
}

// RenderSource writes the given Typst source code to a temp file, compiles it
// with Typst CLI to PNG, and returns a ReadCloser on the image file plus a
// cleanup func.
func RenderSource(source, name string) (io.ReadCloser, func(), error) {
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

	if _, err := typFile.WriteString(source); err != nil {
		typFile.Close()
		cleanup()
		return nil, nil, fmt.Errorf("write %s typ: %w", name, err)
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

// RenderFile compiles an existing .typ file with Typst CLI to PNG, and returns
// a ReadCloser on the image file plus a cleanup func.
func RenderFile(filePath string) (io.ReadCloser, func(), error) {
	pngFile, err := os.CreateTemp("", "doc-*.png")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp png: %w", err)
	}
	pngPath := pngFile.Name()
	pngFile.Close()

	cleanup := func() {
		os.Remove(pngPath)
	}

	cmd := exec.Command("typst", "c", filePath, pngPath, "-f", "png")
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("typst compile %s: %w (%s)", filePath, err, string(out))
	}

	f, err := os.Open(pngPath)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("open rendered png %s: %w", pngPath, err)
	}
	return f, cleanup, nil
}

// RenderSourceToImage compiles Typst source to a cropped image.Image.
func RenderSourceToImage(source, name string) (image.Image, error) {
	pngFile, cleanup, err := RenderSource(source, name)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	defer pngFile.Close()

	img, _, err := image.Decode(pngFile)
	if err != nil {
		return nil, fmt.Errorf("decode rendered %s png: %w", name, err)
	}
	return CropHeight8(img), nil
}

// RenderFileToImage compiles an existing .typ file to a cropped image.Image.
func RenderFileToImage(filePath string) (image.Image, error) {
	pngFile, cleanup, err := RenderFile(filePath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	defer pngFile.Close()

	img, _, err := image.Decode(pngFile)
	if err != nil {
		return nil, fmt.Errorf("decode rendered %s png: %w", filePath, err)
	}
	return CropHeight8(img), nil
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
