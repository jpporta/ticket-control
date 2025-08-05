package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"

	"github.com/hennedo/escpos"
	"github.com/jackc/pgx/v5"
	"github.com/jpporta/ticket-control/internal/repository"
)

type PrinterConfig struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Enabled bool   `json:"enabled"`
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)
	queries := repository.New(conn)
	config, err := queries.GetPrinterConfig(ctx)
	var cfg PrinterConfig
	if err = json.Unmarshal(config, &cfg); err != nil {
		panic(err)
	}
	socket, err := net.Dial("tcp", cfg.IP+":"+strconv.Itoa(cfg.Port))
	if err != nil {
		println(err.Error())
	}
	defer socket.Close()

	p := escpos.New(socket)

	p.HRIFont(true)
	p.Write("Hello true")
	p.LineFeed()
	p.HRIFont(false)
	p.Write("Hello true")
	p.LineFeed()
	p.HRIPosition(0)
	p.Write("Hello 0")
	p.LineFeed()
	p.HRIPosition(1)
	p.Write("Hello 1")
	p.LineFeed()
	p.HRIPosition(2)
	p.Write("Hello 2")
	p.LineFeed()
	p.HRIPosition(3)
	p.Write("Hello 3")
	p.LineFeed()

	// You need to use either p.Print() or p.PrintAndCut() at the end to send the data to the printer.
	p.PrintAndCut()
}
