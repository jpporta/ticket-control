package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hennedo/escpos"
	"github.com/jpporta/ticket-control/internal/clock"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/printer/render"
)

var monthsPT = [...]string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

func formatDatePT(t time.Time) string {
	return fmt.Sprintf("%d de %s de %d", t.Day(), monthsPT[t.Month()-1], t.Year())
}

func (a app) cmdPrint(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "print: missing subcommand (letter, doc, task, list)")
		os.Exit(2)
	}
	switch args[0] {
	case "letter":
		a.printLetter(args[1:])
	case "doc", "typst":
		a.printDoc(args[1:])
	case "task":
		a.printTask(args[1:])
	case "list":
		a.printList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "print: unknown subcommand %q\n", args[0])
		os.Exit(2)
	}
}

func (a app) printLetter(args []string) {
	fs := flag.NewFlagSet("print letter", flag.ExitOnError)
	fileFlag := fs.String("file", "", "path to markdown/text letter file")
	font := fs.String("font", "Libertinus Serif", "font family (e.g. Libertinus Serif, BerkeleyMono Nerd Font Mono, DejaVu Serif)")
	size := fs.String("size", "11pt", "font size (e.g. 10pt, 11pt, 12pt)")
	title := fs.String("title", "", "letter title or header")
	to := fs.String("to", "", "recipient")
	toLabel := fs.String("to-label", "", "recipient header label (default 'Para' in PT, 'To' in EN)")
	from := fs.String("from", "", "sender / signature")
	signOff := fs.String("sign-off", "", "sign-off closing phrase (default 'Atenciosamente,' in PT; pass empty string to omit)")
	lang := fs.String("lang", "pt", "language for headers/date (pt|en, default pt)")
	date := fs.String("date", "", "custom date header (defaults to current date in selected language)")
	noDate := fs.Bool("no-date", false, "omit date from header")
	align := fs.String("align", "justify", "alignment: justify|left|center")
	preview := fs.String("preview", "", "save rendered image to file without printing")
	ip := fs.String("ip", "", "printer IP (default $PRINTER_IP, then config)")
	port := fs.String("port", os.Getenv("PRINTER_PORT"), "printer port (default 9100)")

	parsedArgs, positional := splitFlagsAndPositional(args)
	if err := fs.Parse(parsedArgs); err != nil {
		os.Exit(2)
	}

	filePath := *fileFlag
	if filePath == "" && len(positional) > 0 {
		filePath = positional[0]
	}

	var content string
	if filePath != "" && filePath != "-" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read letter file: %v\n", err)
			os.Exit(1)
		}
		content = string(data)
	} else if !isTerminal(os.Stdin) {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
			os.Exit(1)
		}
		content = string(data)
	} else {
		fmt.Fprintln(os.Stderr, "error: letter file or piped stdin required")
		fmt.Fprintln(os.Stderr, "usage: ticket-control print letter [flags] <file.md>")
		os.Exit(2)
	}

	now := time.Now()
	dateVal := *date
	if *noDate {
		dateVal = ""
	} else if dateVal == "" {
		if *lang == "en" {
			dateVal = now.Format("02 January 2006")
		} else {
			dateVal = formatDatePT(now)
		}
	}

	toLabelVal := *toLabel
	if toLabelVal == "" {
		if *lang == "en" {
			toLabelVal = "To"
		} else {
			toLabelVal = "Para"
		}
	}

	signOffVal := *signOff
	if signOffVal == "" && !isFlagPassed(fs, "sign-off") {
		if *lang == "en" {
			signOffVal = "Warm regards,"
		} else {
			signOffVal = "Atenciosamente,"
		}
	}

	input := printer.LetterInput{
		Title:    *title,
		Date:     dateVal,
		To:       *to,
		ToLabel:  toLabelVal,
		From:     *from,
		SignOff:  signOffVal,
		Font:     *font,
		FontSize: *size,
		Justify:  *align == "justify",
		Content:  content,
	}

	img, err := printer.RenderLetter(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render letter: %v\n", err)
		os.Exit(1)
	}

	if *preview != "" {
		if err := savePNG(*preview, img); err != nil {
			fmt.Fprintf(os.Stderr, "save preview: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("saved preview to %s\n", *preview)
		return
	}

	resolvedIP := a.printerIP(*ip)
	if resolvedIP == "" {
		fmt.Fprintln(os.Stderr, "printer IP required (set in config.toml, PRINTER_IP, or --ip)")
		os.Exit(1)
	}
	if *port == "" {
		*port = "9100"
	}
	addr := net.JoinHostPort(resolvedIP, *port)

	if err := printImageDirect(addr, img); err != nil {
		fmt.Fprintf(os.Stderr, "print: %v\n", err)
		os.Exit(1)
	}
}

func isFlagPassed(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func (a app) printDoc(args []string) {
	fs := flag.NewFlagSet("print doc", flag.ExitOnError)
	fileFlag := fs.String("file", "", "path to .typ file")
	preview := fs.String("preview", "", "save rendered image to file without printing")
	ip := fs.String("ip", "", "printer IP (default $PRINTER_IP, then config)")
	port := fs.String("port", os.Getenv("PRINTER_PORT"), "printer port (default 9100)")

	parsedArgs, positional := splitFlagsAndPositional(args)
	if err := fs.Parse(parsedArgs); err != nil {
		os.Exit(2)
	}

	filePath := *fileFlag
	if filePath == "" && len(positional) > 0 {
		filePath = positional[0]
	}

	var img image.Image
	var err error

	if filePath != "" && filePath != "-" {
		img, err = render.RenderFileToImage(filePath)
	} else if !isTerminal(os.Stdin) {
		data, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "read stdin: %v\n", readErr)
			os.Exit(1)
		}
		img, err = render.RenderSourceToImage(string(data), "doc")
	} else {
		fmt.Fprintln(os.Stderr, "error: .typ file or piped stdin required")
		fmt.Fprintln(os.Stderr, "usage: ticket-control print doc [flags] <file.typ>")
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "render doc: %v\n", err)
		os.Exit(1)
	}

	if *preview != "" {
		if err := savePNG(*preview, img); err != nil {
			fmt.Fprintf(os.Stderr, "save preview: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("saved preview to %s\n", *preview)
		return
	}

	resolvedIP := a.printerIP(*ip)
	if resolvedIP == "" {
		fmt.Fprintln(os.Stderr, "printer IP required (set in config.toml, PRINTER_IP, or --ip)")
		os.Exit(1)
	}
	if *port == "" {
		*port = "9100"
	}
	addr := net.JoinHostPort(resolvedIP, *port)

	if err := printImageDirect(addr, img); err != nil {
		fmt.Fprintf(os.Stderr, "print: %v\n", err)
		os.Exit(1)
	}
}

func (a app) printTask(args []string) {
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

func (a app) printList(args []string) {
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

func splitFlagsAndPositional(args []string) ([]string, []string) {
	var flags, positional []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				if isValueFlag(arg) {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			positional = append(positional, arg)
		}
		i++
	}
	return flags, positional
}

func isValueFlag(flagName string) bool {
	name := strings.TrimLeft(flagName, "-")
	switch name {
	case "file", "font", "size", "title", "to", "to-label", "from", "sign-off", "signoff", "lang", "date", "align", "preview", "ip", "port":
		return true
	default:
		return false
	}
}

func printImageDirect(addr string, img image.Image) error {
	var conn net.Conn
	var err error
	for attempt := range 3 {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
		if attempt < 2 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("printer unreachable at %s: %w", addr, err)
	}
	defer conn.Close()

	e := escpos.New(conn)
	conn.Write([]byte{0x1B, 0x40, 0x1B, 0x52, 0x00})
	if _, err := e.PrintImage(img); err != nil {
		return fmt.Errorf("send image to printer: %w", err)
	}
	return e.PrintAndCut()
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
