package task

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jpporta/ticket-control/printer"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func PrintImage(ctx context.Context, img []byte) error {
	start := time.Now()
	defer func() {
		log.Printf("PrintImage took %s", time.Since(start))
	}()

	printer_port := os.Getenv("PRINTER_PORT")
	if printer_port == "" {
		return fmt.Errorf("PRINTER_PORT environment variable not set")
	}
	conn, err := grpc.NewClient(":"+printer_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("error creating gRPC client: %w", err)
	}
	defer conn.Close()

	pr := printer.NewPrinterClient(conn)

	job := printer.PrintJob{
		Img: img,
	}
	_, err = pr.Print(ctx, &job)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}
