package main

import (
	"context"
	"image"
	"image/draw"
	_ "image/jpeg"
	"os"

	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/utils"
)

func main() {
	printTaskTest()
}

func simpleTest() {
	ctx := context.Background()

	// Printer
	p := printer.New(ctx)
	close, err := p.Start()
	defer close()

	if err != nil {
		panic(err)
	}
	err = p.PrintText("Hello World\n")
	if err != nil {
		panic(err)
	}
}

func printTaskTest() {
	ctx := context.Background()

	// Printer
	p := printer.New(ctx)
	close, err := p.Start()
	defer close()

	if err != nil {
		panic(err)
	}

	err = p.PrintTask("Test Task", "This is a test task description", 3, "John Doe")
	if err != nil {
		panic(err)
	}
}

func printImageTest() {
ctx := context.Background()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	reader, err := os.Open(home + "/Pictures/gu.jpg")
	defer func() {
		err := reader.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	// Printer
	p := printer.New(ctx)
	close, err := p.Start()
	defer close()

	if err != nil {
		panic(err)
	}

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	draw.Draw(grayImg, bounds, img, bounds.Min, draw.Src)
	w, h := bounds.Max.X, bounds.Max.Y
	d := &utils.Dither{
		SourceImage: grayImg,
		Width:       w,
		Height:      h,
		NewImage:    image.NewRGBA(bounds),
	}

	d.OrderedDither16()
	p.PrintImage(d.NewImage)
	p.Cut()
}
