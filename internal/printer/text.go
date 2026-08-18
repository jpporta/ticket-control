package printer

import (
	"fmt"
	"log/slog"
	"text/template"
)

func (p *Printer) PrintText(text string) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintText(text)
		})
		return fmt.Errorf("%w: queuing text: %s", errPrinterOffline, text)
	}
	close, err := p.start()
	if err != nil {
		return err
	}
	defer close()
	if _, err := p.e.Write(text); err != nil {
		return err
	}
	if err := p.e.PrintAndCut(); err != nil {
		slog.Error("print and cut", "err", err)
		return err
	}
	return nil
}

func (p *Printer) Cut() {
	p.e.WriteRaw([]byte{0x1B, 0x6D})
}

func (p *Printer) loadTemplates() error {
	p.templates = make(map[string]*template.Template)
	for _, name := range []string{"task", "list", "link_header", "end_of_day", "letter"} {
		raw, err := models.ReadFile("models/" + name + ".typ")
		if err != nil {
			return err
		}
		t, err := template.New(name).Parse(string(raw))
		if err != nil {
			return err
		}
		p.templates[name] = t
	}
	return nil
}
