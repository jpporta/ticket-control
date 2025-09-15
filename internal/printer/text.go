package printer

import (
	"fmt"
	"text/template"
)

func (p *Printer) PrintText(text string) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintText(text)
		})
		return fmt.Errorf("Printer is disabled, queuing text: %s\n", text)
	}
	close, err := p.start()
	if err != nil {
		return err
	}
	defer close()
	_, err = p.e.Write(text)
	if err != nil {
		return err
	}
	err = p.e.PrintAndCut()
	if err != nil {
		return err
	}
	return nil
}

func (p *Printer) Cut() {
	p.e.WriteRaw([]byte{0x1B, 0x6D})
}

func (p *Printer) loadTemplates() error {
	p.templates = make(map[string]*template.Template)
	// Task template
	task_template_string, err := models.ReadFile("models/task.typ")
	if err != nil {
		return err
	}
	task_template, err := template.New("task").Parse(string(task_template_string))
	if err != nil {
		return err
	}
	p.templates["task"] = task_template
	// List template
	list_template_string, err := models.ReadFile("models/list.typ")
	if err != nil {
		return err
	}
	list_template, err := template.New("list").Parse(string(list_template_string))
	if err != nil {
		return err
	}
	p.templates["list"] = list_template
	// Link Header template
	link_header_template_string, err := models.ReadFile("models/link_header.typ")
	if err != nil {
		return err
	}
	link_header_template, err := template.New("list").Parse(string(link_header_template_string))
	if err != nil {
		return err
	}
	p.templates["link_header"] = link_header_template

	// End of day template
	end_of_day_template_string, err := models.ReadFile("models/end_of_day.typ")
	if err != nil {
		return err
	}
	end_of_day_template, err := template.New("end_of_day").Parse(string(end_of_day_template_string))
	if err != nil {
		return err
	}
	p.templates["end_of_day"] = end_of_day_template
	return nil
}

