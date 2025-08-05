package main

import (
	"context"
	"image"
	"image/draw"
	"os"

	_ "image/jpeg"

	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/utils"
)

func main() {
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
}
