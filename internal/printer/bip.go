package printer

import "fmt"

func (p *Printer) PrintBip() error {
	if !p.Enabled {
		return fmt.Errorf("Printer is disabledi, ignoring task\n")
	}
	close, err := p.start()
	if err != nil {
		return err
	}
	defer close()
	p.e.WriteRaw([]byte{0x1B, 0x69})
	p.e.Print()
	return nil
}
