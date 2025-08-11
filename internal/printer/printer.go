package printerInternal

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image"
	"image/png"
	"text/template"

	"github.com/jpporta/ticket-control/printer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed models/*.typ
var models embed.FS

type Printer struct {
	templates map[string]*template.Template
}

func New() (*Printer, error) {
	p := &Printer{}
	if err := p.loadTemplates(); err != nil {
		return nil, err
	}
	return p, nil
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
	return nil
}

func (p *Printer) PrintImage(img image.Image) error {
	conn, err := grpc.NewClient(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("error creating gRPC client: %w", err)
	}
	defer conn.Close()

	pr := printer.NewPrinterClient(conn)

	img_buf := new(bytes.Buffer)
	err = png.Encode(img_buf, img)
	job := printer.PrintJob{
		Img: img_buf.Bytes(),
	}
	_, err = pr.Print(context.Background(), &job)

	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}

func (p *Printer) PrintLinkCall(header image.Image, url string) error {
	conn, err := grpc.NewClient(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("error creating gRPC client: %w", err)
	}
	defer conn.Close()

	pr := printer.NewPrinterClient(conn)

	img_buf := new(bytes.Buffer)
	err = png.Encode(img_buf, header)
	job := printer.PrintLinkJob{
		Header: img_buf.Bytes(),
		Url: url,
	}
	_, err = pr.PrintLink(context.Background(), &job)

	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}
