package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jpporta/ticket-control/task"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	port := os.Getenv("TASK_PORT")
	if port == "" {
		log.Fatalf("TASK_PORT environment variable is not set")
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()
	conn, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
	s := task.NewServer(conn)

	printerGRPCServer := grpc.NewServer()
	task.RegisterTaskServiceServer(printerGRPCServer, s)

	if err := printerGRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
