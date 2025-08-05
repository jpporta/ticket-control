package printer

import "image"

func (p *Printer) PrintImage(img image.Image) {
	p.e.PrintImage(img)
	p.e.PrintAndCut()
}
