package printer

import (
	"fmt"
	"image"

	"github.com/jpporta/ticket-control/internal/utils"
)

func (p *Printer) PrintImage(img image.Image) error {
	data,x,y, err := utils.ImageToBytes(img)
	fmt.Println("Image size:", x, y, "Data length:", len(data))
	fmt.Println(byte(x >> 3) & 0xff,
		byte(x >> 11) & 0xff, byte(y & 0xff), byte(y >> 8) & 0xff)

	p.e.WriteRaw([]byte{0x1D, 0x76, 0x30, 0x00, byte(x >> 3) & 0xff,
		byte(x >> 11) & 0xff, byte(y & 0xff), byte(y >> 8) & 0xff})
	p.e.WriteRaw(data)
	p.e.PrintAndCut()
	return err
}
