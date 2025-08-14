package printer

import (
	"context"
	"embed"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"text/template"

	"github.com/hennedo/escpos"
	"github.com/jackc/pgx/v5"
	"github.com/jpporta/ticket-control/internal/repository"
)

//go:embed models/*.typ
var models embed.FS

type Printer struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Enabled   bool   `json:"enabled"`
	e         *escpos.Escpos
	templates map[string]*template.Template
}

func New(ctx context.Context) *Printer {
	conn, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)
	queries := repository.New(conn)
	config, err := queries.GetPrinterConfig(ctx)
	printer := &Printer{}

	if err = json.Unmarshal(config, printer); err != nil {
		panic(err)
	}
	return printer
}

func (p *Printer) Start() (func(), error) {
	err := p.loadTemplates()
	if err != nil {
		return nil, err
	}

	socket, err := net.Dial("tcp", p.IP+":"+strconv.Itoa(p.Port))
	if err != nil {
		return nil, err
	}

	p.e = escpos.New(socket)
	return func() {
		err := socket.Close()
		if err != nil {
			panic(err)
		}
	}, nil
}

func (p *Printer) Reset() {
	p.e.WriteRaw([]byte{0x1B, byte('@')})
	p.e.WriteRaw([]byte{0x1B, 0x52, 0x00})
}

func (p *Printer) PrintText(text string) error {
	if !p.Enabled {
		return nil
	}
	_, err := p.e.Write(text)
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
