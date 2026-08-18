package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// ESC/POS control sequences. Only the handful this client needs; the escpos
// library is avoided here because it pulls in cgo iconv for charset handling we
// do ourselves with x/text.
var (
	escReset    = []byte{0x1B, 0x40}       // ESC @
	escFeed5    = []byte{0x1B, 0x64, 0x05} // ESC d 5 — feed past the cutter
	escCut      = []byte{0x1B, 0x6D}       // ESC m — partial cut
	escCodepage = func(id byte) []byte { return []byte{0x1B, 0x74, id} }
)

// escDisambiguation is how long we wait after a bare ESC byte before deciding
// it was a real Esc keypress rather than the prefix of an arrow-key sequence.
const escDisambiguation = 50 * time.Millisecond

type codepage struct {
	id  byte
	enc encoding.Encoding
}

var codepages = map[string]codepage{
	"cp437":   {0, charmap.CodePage437},
	"cp850":   {2, charmap.CodePage850},
	"cp860":   {3, charmap.CodePage860},
	"win1252": {16, charmap.Windows1252},
}

type session struct {
	addr string
	cp   codepage
	line []rune
}

// dial opens a connection and puts the printer into a known state.
func (s *session) dial() (net.Conn, error) {
	c, err := net.Dial("tcp", s.addr)
	if err != nil {
		return nil, err
	}
	init := append(append([]byte{}, escReset...), escCodepage(s.cp.id)...)
	if _, err := c.Write(init); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// write dials, writes and closes. The printer's TCP port accepts a single
// client, so holding it open for a whole session would lock the server out of
// its own printer; a connection per line keeps that window at milliseconds.
func (s *session) write(b []byte) error {
	c, err := s.dial()
	if err != nil {
		return err
	}
	defer c.Close()
	_, err = c.Write(b)
	return err
}

// encode transcodes text to the selected codepage. Runes with no representation
// are dropped rather than emitted as an incorrect byte.
func (s *session) encode(text string) []byte {
	enc := s.cp.enc.NewEncoder()
	out := make([]byte, 0, len(text))
	for _, r := range text {
		b, err := enc.Bytes([]byte(string(r)))
		if err != nil {
			continue
		}
		out = append(out, b...)
	}
	return out
}

func (s *session) cut() error {
	return s.write(append(append([]byte{}, escFeed5...), escCut...))
}

// commit sends the buffered line, terminated by LF — which is what actually
// makes the print head fire.
func (s *session) commit() error {
	out := append(s.encode(string(s.line)), '\n')
	s.line = s.line[:0]
	return s.write(out)
}

func (a app) typewriter(args []string) {
	fs := flag.NewFlagSet("typewriter", flag.ExitOnError)
	ip := fs.String("ip", "", "printer IP (default $PRINTER_IP, then config)")
	port := fs.String("port", os.Getenv("PRINTER_PORT"), "printer port (default $PRINTER_PORT)")
	cpName := fs.String("codepage", "cp850", "cp437|cp850|cp860|win1252")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	*ip = a.printerIP(*ip)
	if *port == "" {
		*port = "9100"
	}
	if _, err := strconv.Atoi(*port); err != nil {
		fmt.Fprintf(os.Stderr, "invalid port %q\n", *port)
		os.Exit(2)
	}
	cp, ok := codepages[*cpName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown codepage %q\n", *cpName)
		os.Exit(2)
	}

	s := &session{addr: net.JoinHostPort(*ip, *port), cp: cp}

	// Fail before touching the terminal if the printer is unreachable.
	if c, err := s.dial(); err != nil {
		fmt.Fprintf(os.Stderr, "printer unreachable: %v\n", err)
		os.Exit(1)
	} else {
		c.Close()
	}

	// Piped or redirected stdin: print it and exit. `cal | tc typewriter`.
	if !isTerminal(os.Stdin) {
		if err := s.pipe(os.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("typewriter · %s · %s\n", s.addr, *cpName)
	fmt.Println("Enter commit · Esc cut+continue · Ctrl-D cut+exit · Ctrl-C abort")

	restore, err := rawTerminal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal: %v\n", err)
		os.Exit(1)
	}
	defer restore()

	// Ctrl-C does not raise SIGINT in raw mode (isig is off) — we read 0x03
	// ourselves. This handler covers SIGTERM/SIGHUP so the terminal is never
	// left without echo.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigCh
		restore()
		os.Exit(1)
	}()

	if err := s.loop(readRunes(os.Stdin)); err != nil {
		restore()
		fmt.Fprintf(os.Stderr, "\n%v\n", err)
		os.Exit(1)
	}
}

// pipe prints everything on r as one ticket, over a single connection.
func (s *session) pipe(r io.Reader) error {
	text, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	c, err := s.dial()
	if err != nil {
		return err
	}
	defer c.Close()
	out := s.encode(string(text))
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	out = append(out, escFeed5...)
	out = append(out, escCut...)
	_, err = c.Write(out)
	return err
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// readRunes decodes stdin into runes on a channel so the main loop can apply a
// timeout when disambiguating a bare ESC from an escape sequence.
func readRunes(f *os.File) <-chan rune {
	ch := make(chan rune)
	go func() {
		defer close(ch)
		br := bufio.NewReader(f)
		for {
			r, _, err := br.ReadRune()
			if err != nil {
				return
			}
			ch <- r
		}
	}()
	return ch
}

func (s *session) loop(in <-chan rune) error {
	for r := range in {
		switch r {
		case 0x03: // Ctrl-C — abort, deliberately no cut
			fmt.Print("\r\n")
			return nil

		case 0x04: // Ctrl-D — cut and exit
			if err := s.flush(); err != nil {
				return err
			}
			if err := s.cut(); err != nil {
				return err
			}
			fmt.Print("\r\n")
			return nil

		case 0x1B: // Esc, or the start of an escape sequence
			if isEscKey(in) {
				if err := s.flush(); err != nil {
					return err
				}
				if err := s.cut(); err != nil {
					return err
				}
				fmt.Print("\r\n--- cut ---\r\n")
			}

		case '\r', '\n':
			fmt.Print("\r\n")
			if err := s.commit(); err != nil {
				return err
			}

		case 0x7F, 0x08: // Backspace
			if len(s.line) > 0 {
				s.line = s.line[:len(s.line)-1]
				fmt.Print("\b \b")
			}

		default:
			if r < 0x20 {
				continue // unhandled control byte
			}
			s.line = append(s.line, r)
			fmt.Print(string(r))
		}
	}
	return nil
}

// isEscKey drains an escape sequence if one follows. Returns true when the ESC
// stood alone, i.e. the user actually pressed Esc.
func isEscKey(in <-chan rune) bool {
	t := time.NewTimer(escDisambiguation)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case _, ok := <-in:
		if !ok {
			return true
		}
		// An escape sequence; discard the remainder of it.
		for {
			t.Reset(escDisambiguation)
			select {
			case <-t.C:
				return false
			case _, ok := <-in:
				if !ok {
					return false
				}
			}
		}
	}
}

// flush commits any pending line so it is not lost on cut or exit.
func (s *session) flush() error {
	if len(s.line) == 0 {
		return nil
	}
	fmt.Print("\r\n")
	return s.commit()
}

// rawTerminal puts the terminal into raw mode using stty, avoiding a dependency
// on golang.org/x/term for what is two shell calls.
// ponytail: stty, swap for x/term if this ever needs to run on Windows.
func rawTerminal() (func(), error) {
	saved, err := stty("-g")
	if err != nil {
		return nil, err
	}
	if _, err := stty("raw", "-echo"); err != nil {
		return nil, err
	}
	var once bool
	return func() {
		if once {
			return
		}
		once = true
		stty(saved)
	}, nil
}

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("stty %v: %w", args, err)
	}
	return string(trimNewline(out)), nil
}

func trimNewline(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}
