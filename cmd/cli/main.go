package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
	"github.com/jpporta/ticket-control/internal/utils"
)

const usage = `ticket-control CLI

Subcommands:
  user create --name <name>                           Create a user, print API key
  printer test                                        Send a test "bip" to the printer
  print letter [flags] <file.md>                      Print a formatted letter from Markdown/text
  print doc [flags] <file.typ>                        Print a Typst document directly
  print task --title T --description D --priority P   Print a task ticket
  print list --title T --items "a,b,c"                Print a list ticket
  typewriter [--ip IP] [--port P] [--codepage CP]     Type straight onto the printer.
                                                      With piped stdin, prints it and cuts.

Run a subcommand with --help for flag details.
`

func main() {
	app, err := newApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		if !isTerminal(os.Stdin) {
			app.typewriter(os.Args[1:])
			return
		}
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "typewriter":
		app.typewriter(os.Args[2:])
	case "user":
		cmdUser(os.Args[2:])
	case "printer":
		cmdPrinter(os.Args[2:])
	case "print":
		app.cmdPrint(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		if strings.HasPrefix(os.Args[1], "-") {
			app.typewriter(os.Args[1:])
			return
		}
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
