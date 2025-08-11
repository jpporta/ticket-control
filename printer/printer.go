package printer

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"net"
	"os"
	"strconv"

	"github.com/hennedo/escpos"
	"github.com/jackc/pgx/v5"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Printer struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Enabled   bool   `json:"enabled"`
	e         *escpos.Escpos
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

func (p *Printer) PrintImage(img image.Image) error {
	_, err := p.e.PrintImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return p.e.PrintAndCut()
}
