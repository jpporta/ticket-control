package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jpporta/ticket-control/internal"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
	"github.com/jpporta/ticket-control/internal/utils"
)

func main() {
	printListTest()
}

func printListTest() {
	ctx := context.WithValue(context.Background(), "userName", "Joao Porta Testing")
	conn, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())
	app := internal.New(conn)
	_, err = app.CreateList(ctx, 2, "Testing List", []string{
		"Item 1",
		"Item 2",
		"Item 3",
		"Item 4",
		"Item 5",
		"Item 6",
		"Item 7",
		"Item muito longo para poder wrap de linha e usar duas linhas para testar o template",
	})
	if err != nil {
		panic(err)
	}
}

func createUser() {
	name := flag.String("name", "", "Name of the user")
	flag.Parse()

	if *name == "" {
		panic("Name is required")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())
	app := internal.New(conn)
	key := utils.RandomString(32)
	_, err = app.Q.CreateUser(ctx, repository.CreateUserParams{
		Name:   *name,
		ApiKey: key,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("User created successfully", key)
}

func simpleTest() {
	ctx := context.Background()

	// Printer
	p := printer.New(ctx)
	close, err := p.Start()
	defer close()

	if err != nil {
		panic(err)
	}
	err = p.PrintText("Hello World\n")
	if err != nil {
		panic(err)
	}
}

func printTaskTest() {
	ctx := context.Background()

	// Printer
	p := printer.New(ctx)
	close, err := p.Start()
	defer close()

	if err != nil {
		panic(err)
	}

	err = p.PrintTask("Test Task", "This is a test task description", 3, "John Doe")
	if err != nil {
		panic(err)
	}
}

func printImageTest() {
	ctx := context.Background()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	reader, err := os.Open(home + "/Pictures/gu.jpg")
	defer func() {
		err := reader.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	// Printer
	p := printer.New(ctx)
	close, err := p.Start()
	defer close()

	if err != nil {
		panic(err)
	}

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	draw.Draw(grayImg, bounds, img, bounds.Min, draw.Src)
	w, h := bounds.Max.X, bounds.Max.Y
	d := &utils.Dither{
		SourceImage: grayImg,
		Width:       w,
		Height:      h,
		NewImage:    image.NewRGBA(bounds),
	}

	d.OrderedDither16()
	p.PrintImage(d.NewImage)
	p.Cut()
}
