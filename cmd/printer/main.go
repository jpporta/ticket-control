package main

import (
	"log"
	"net"

	"github.com/jpporta/ticket-control/printer"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()
	s := printer.Server{}

	printerGRPCServer := grpc.NewServer()
	printer.RegisterPrinterServer(printerGRPCServer, &s)

	if err := printerGRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
