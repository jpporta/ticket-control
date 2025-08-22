package printer

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"strings"
)

type ListInput struct {
	Title     string
	Content   []string
	CreatedBy string
}

func (p *Printer) PrintList(
	list ListInput,
) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintList(list)
		})
		return fmt.Errorf("Printer is disabled, queuing list: %s\n", list.Title)
	}

	// Load Template
	template, ok := p.templates["list"]
	if !ok {
		return fmt.Errorf("list template not found")
	}

	// Create temporary file for Typst
	file, err := os.CreateTemp("", "list-*.typ")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}

	// Execute the template into the temporary file
	template.Execute(file, list)

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

	close, err := p.start()
	if err != nil {
		fmt.Println("Error starting printer:", err)
		return err
	}
	defer close()
	// Reset the printer state
	p.Reset()
	err = p.printImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}
