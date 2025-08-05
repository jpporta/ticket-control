package utils

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
)

type Dither struct {
	SourceImage   *image.Gray // pointer to the source image in grayscale
	Width, Height int         // dimensions of the source image
	NewImage      draw.Image  // resulting dithered image
	Threshold     int         // threshold value used in dithering algorithm
}

func (d *Dither) GrayDither() {
	// sections := []int{85, 170, 255}
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			px := d.getPixel(col, row)
			// if px < sections[0] {
			// 	d.NewImage.Set(col, row, color.Black)
			// 	continue
			// }
			// if px < sections[1] {
			// 	d.NewImage.Set(col, row, color.Gray{Y: uint8(sections[1])})
			// 	continue
			// }
			d.NewImage.Set(col, row, color.Gray{Y: uint8(px)})
		}
	}
}

// Default dithering
func (d *Dither) OrderedDither4() {
	dots := [][]int{{64, 128}, {192, 0}}
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			dotrow := 1
			if row%2 == 0 {
				dotrow = 0
			}
			dotcol := 1
			if col%2 == 0 {
				dotcol = 0
			}
			px := d.getPixel(col, row)
			if px > dots[dotrow][dotcol] {
				d.NewImage.Set(col, row, color.White)
			} else {
				d.NewImage.Set(col, row, color.Black)
			}
		}
	}
}

func (d *Dither) OrderedDither9() {
	dots := [][]int{{0, 196, 84}, {168, 140, 56}, {112, 28, 224}}
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			dotrow := 0
			if row%3 == 0 {
				dotrow = 2
			} else if row%2 == 0 {
				dotrow = 1
			}
			dotcol := 0
			if col%3 == 0 {
				dotcol = 2
			} else if col%2 == 0 {
				dotcol = 1
			}
			px := d.getPixel(col, row)
			if px > dots[dotrow][dotcol] {
				d.NewImage.Set(col, row, color.White)
			} else {
				d.NewImage.Set(col, row, color.Black)
			}
		}
	}
}

func (d *Dither) OrderedDither16() {
	dots := [][]int{{0, 196, 84, 15}, {168, 140, 56, 12}, {112, 28, 224, 128}}
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			px := d.getPixel(col, row)
			if px > dots[row%3][col%4] {
				d.NewImage.Set(col, row, color.White)
			} else {
				d.NewImage.Set(col, row, color.Black)
			}
		}
	}
}

func (d *Dither) ThresholdDither() {
	if d.Threshold == 0 {
		pxList := d.SourceImage.Pix
		d.Threshold = 0
		for i := range pxList {
			d.Threshold += int(pxList[i])
		}
		d.Threshold = d.Threshold / len(pxList)
	}
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			px := d.getPixel(col, row)
			if px > d.Threshold {
				d.NewImage.Set(col, row, color.White)
			} else {
				d.NewImage.Set(col, row, color.Black)
			}
		}
	}
}

func (d *Dither) RandomDither() {
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			px := d.getPixel(col, row)
			rand := rand.Intn(255)
			if px > rand {
				d.NewImage.Set(col, row, color.White)
			} else {
				d.NewImage.Set(col, row, color.Black)
			}
		}
	}
}

func (d *Dither) getPixel(x int, y int) int {
	if x > d.Width || y > d.Height {
		return 0
	}
	r, g, b, _ := d.SourceImage.At(x, y).RGBA()
	res := uint8((r + g + b) / 3)
	return int(res)
}
