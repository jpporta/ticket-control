package printerInternal

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"strings"
	"time"
	_ "image/png"
)

type taskInput struct {
	Title           string
	Description     string
	PriorityDisplay string
	CreatedBy       string
	CreatedAt       time.Time
}

func (p *Printer) PrintTask(
	title, description string,
	priority int32,
	createdBy string,
) error {
	template, ok := p.templates["task"]
	if !ok {
		return fmt.Errorf("task template not found")
	}

	file, err := os.CreateTemp("", "task-*.typ")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	template.Execute(file, taskInput{
		Title:           title,
		Description:     description,
		PriorityDisplay: strings.TrimSpace(strings.Repeat(" ", int(priority))),
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
	})
	cmd := exec.Command("typst", "c", file.Name(), "-f", "png")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing typst command: %w", err)
	}

	img_raw, err := os.Open(strings.Replace(file.Name(), ".typ", ".png", 1))
	if err != nil {
		return fmt.Errorf("error opening image file: %w", err)
	}
	defer img_raw.Close()
	img, _, err := image.Decode(img_raw)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}
	if img.Bounds().Max.Y%8 != 0 {
		cropRect := image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y-(img.Bounds().Max.Y%8))
		img = img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(cropRect)
	}
	err = p.PrintImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}

