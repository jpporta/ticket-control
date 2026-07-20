package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
	"github.com/jpporta/ticket-control/internal/utils"
)

const usage = `ticket-control CLI

Subcommands:
  user create --name <name>      Create a user, print API key
  printer test                   Send a test "bip" to the printer
  print task --title T --description D --priority P   Print a task ticket
  print list --title T --items "a,b,c"                Print a list ticket

Run a subcommand with --help for flag details.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "user":
		cmdUser(os.Args[2:])
	case "printer":
		cmdPrinter(os.Args[2:])
	case "print":
		cmdPrint(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

func cmdUser(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "user: missing subcommand")
		os.Exit(2)
	}
	switch args[0] {
	case "create":
		userCreate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "user: unknown subcommand %q\n", args[0])
		os.Exit(2)
	}
}

func cmdPrinter(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "printer: missing subcommand")
		os.Exit(2)
	}
	switch args[0] {
	case "test":
		printerTest(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "printer: unknown subcommand %q\n", args[0])
		os.Exit(2)
	}
}

func cmdPrint(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "print: missing subcommand")
		os.Exit(2)
	}
	switch args[0] {
	case "task":
		printTask(args[1:])
	case "list":
		printList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "print: unknown subcommand %q\n", args[0])
		os.Exit(2)
	}
}

func mustPool() *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	return conn
}

func userCreate(args []string) {
	fs := flag.NewFlagSet("user create", flag.ExitOnError)
	name := fs.String("name", "", "name of the user")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *name == "" {
		fmt.Fprintln(os.Stderr, "--name is required")
		os.Exit(2)
	}
	ctx := context.Background()
	conn := mustPool()
	defer conn.Close()
	q := repository.New(conn)
	key := utils.RandomString(32)
	id, err := q.CreateUser(ctx, repository.CreateUserParams{Name: *name, ApiKey: key})
	if err != nil {
		panic(err)
	}
	fmt.Printf("user created id=%d api_key=%s\n", id, key)
}

func printerTest(args []string) {
	fs := flag.NewFlagSet("printer test", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	p := printer.New(context.Background())
	if err := p.PrintBip(); err != nil {
		panic(err)
	}
}

func printTask(args []string) {
	fs := flag.NewFlagSet("print task", flag.ExitOnError)
	title := fs.String("title", "", "task title")
	desc := fs.String("description", "", "task description")
	prio := fs.Int("priority", 0, "task priority (-2..5)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *title == "" {
		fmt.Fprintln(os.Stderr, "--title is required")
		os.Exit(2)
	}
	p := printer.New(context.Background())
	if err := p.PrintTask(0, *title, *desc, int32(*prio), "CLI", clock.Now()); err != nil {
		panic(err)
	}
}

func printList(args []string) {
	fs := flag.NewFlagSet("print list", flag.ExitOnError)
	title := fs.String("title", "", "list title")
	items := fs.String("items", "", "comma-separated items")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *title == "" || *items == "" {
		fmt.Fprintln(os.Stderr, "--title and --items are required")
		os.Exit(2)
	}
	p := printer.New(context.Background())
	if err := p.PrintList(printer.ListInput{
		Title:     *title,
		Content:   strings.Split(*items, ","),
		CreatedBy: "CLI",
	}); err != nil {
		panic(err)
	}
}

// print image: removed; see task 9.1 (legacy implementation depended on
// internal/utils/dither.go, which has been deleted).
