package printerInternal

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"strings"
)

type LinkInput struct {
	Title     string
	URL       string
	CreatedBy string
}

func (p *Printer) PrintLink(
	link LinkInput,
) error {
	// Load Template
	template, ok := p.templates["link_header"]
	if !ok {
		return fmt.Errorf("link template not found")
	}

	// Create temporary file for Typst
	file, err := os.CreateTemp("", "link-*.typ")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}

	// Execute the template into the temporary file
	template.Execute(file, link)

	// Execute Typst command to convert .typ to .png
	cmd := exec.Command("typst", "c", file.Name(), "-f", "png")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing typst command: %w", err)
	}

	// Open the generated image file
	img_raw, err := os.Open(strings.Replace(file.Name(), ".typ", ".png", 1))
	if err != nil {
		return fmt.Errorf("error opening image file: %w", err)
	}
	defer img_raw.Close()

	// Decode the image
	img, _, err := image.Decode(img_raw)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}

	// Crop the image if its height is not a multiple of 8 for the printer
	if img.Bounds().Max.Y%8 != 0 {
		cropRect := image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y-(img.Bounds().Max.Y%8))
		img = img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(cropRect)
	}

	err = p.PrintLinkCall(img, link.URL)

	if err != nil {
		return fmt.Errorf("error printing: %w", err)
	}
	return nil
}
