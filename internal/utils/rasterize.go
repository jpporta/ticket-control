package utils

import (
	"fmt"
	"image"
)

func ImageToBytes(img image.Image) ([]byte, int, int, error) {
	if img == nil {
		return nil, 0, 0, nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	maxW := bounds.Max.X
	maxH := bounds.Max.Y

	data := make([]byte, width*height/8)
	fmt.Println("size:", maxH, maxW)
	for y := range maxH {
		for x := range maxW {
			idx := (y*maxW + x) / 8
			offset := (y*maxW + x) % 8
			r, _, _, _ := img.At(x, y).RGBA()
			if r > 1000 {
				continue
			}
			data[idx] |= (0xff >> offset)
			if offset == 7 {
				println("Pixel at", x, y, "is", r)
				fmt.Println(data[0])
			}
		}
	}

	return data, maxW, maxH, nil
}
