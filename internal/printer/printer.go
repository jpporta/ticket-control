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
	queue     []func() error
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
	err = printer.loadTemplates()
	if err != nil {
		return nil
	}

	if err = json.Unmarshal(config, printer); err != nil {
		panic(err)
	}
	return printer
}

func (p *Printer) start() (func(), error) {
	socket, err := net.Dial("tcp", p.IP+":"+strconv.Itoa(p.Port))
	if err != nil {
		return nil, err
	}

	p.e = escpos.New(socket)
	return func() {
		err := socket.Close()
		p.e = nil
		if err != nil {
			panic(err)
		}
	}, nil
}

func (p *Printer) Reset() {
	p.e.WriteRaw([]byte{0x1B, byte('@')})
	p.e.WriteRaw([]byte{0x1B, 0x52, 0x00})
}
