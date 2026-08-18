# Design

## Shape

```
cmd/cli/typewriter.go
  ├─ flags: --ip --port --codepage (env fallback PRINTER_IP/PRINTER_PORT)
  ├─ terminal raw mode via `stty raw -echo` (no new dependency)
  ├─ keystroke loop over os.Stdin, one line buffered at a time
  ├─ non-tty stdin: print it all as one ticket and cut
  └─ raw ESC/POS bytes written straight to a net.Conn

internal/printer/printer.go
  └─ start(): retry the dial 3× / 2s
```

No `internal/` domain, no port change, no migration, no endpoint, no Typst template.

## Why direct-to-printer

The client is always on the printer's LAN. A server hop would only buy auth,
quota, and persistence — all three are explicit non-goals. So the server is not
in the path at all.

## Hardware constraints that drove the design

1. **ESC/POS commits on `LF`.** Bytes sent without a newline sit in the printer's
   buffer; the head does not move. A per-keystroke mode was implemented and
   tested on the hardware and produced output indistinguishable from the
   buffered path — the only observable difference was the loss of backspace. It
   was removed rather than kept as a mode that costs editing and buys nothing.

2. **Port 9100 is single-client.** While a session holds the socket, the server's
   `net.Dial` fails, the print fails, and the service rolls back the row it just
   inserted — a task silently vanishes. Two mitigations:
   - The client dials per line and closes immediately (~30ms held, not minutes).
   - `printer.start()` retries the dial, covering the collision window and every
     other transient dial failure.

3. **The cutter sits above the print head.** Every cut feeds blank lines first
   (`ESC d 5`) or it slices through the last line of text.

## Encoding

Default codepage CP850 (`ESC t 2`), which carries Portuguese accents. UTF-8 from
the terminal is transcoded with `golang.org/x/text/encoding/charmap` (already an
indirect dependency). Runes with no mapping are dropped rather than emitted as
garbage bytes.

`--codepage` selects among CP437/CP850/CP860/Windows-1252 so the correct one can
be found empirically on the hardware without a code change. This replaces what
would otherwise have been a blocking spike.

The `escpos` library is deliberately not used here — it pulls in cgo `iconv`, and
this client needs only a handful of raw byte sequences.

## Key bindings

| key | byte | action |
|---|---|---|
| Enter | `0x0D` | commit line (sent as `0x0A`) |
| Backspace | `0x7F` | edit buffer (buffered mode only) |
| Esc | `0x1B` | feed + cut, session continues |
| Ctrl-D | `0x04` | feed + cut, close, exit |
| Ctrl-C | `0x03` | abort, **no cut**, restore terminal, exit |

`Ctrl-Enter` was considered and rejected: terminals send `0x0D` for both `Enter`
and `Ctrl-Enter`, so it is indistinguishable outside terminals implementing the
kitty keyboard protocol.

`Esc` is ambiguous with the prefix of arrow-key escape sequences (`ESC [ A`).
Disambiguated by a ~50ms read timeout: a bare `ESC` with nothing following is a
real Esc press; anything following is a sequence and is discarded.

## Echo

Terminal raw mode disables echo, so the client redraws the pending line itself —
required, since you cannot edit a line you cannot see.

## Terminal restore

`stty raw -echo` must be undone on every exit path: normal return, `Ctrl-C`,
panic, and `SIGTERM`/`SIGHUP`. A `defer` alone is insufficient; a signal handler
restores too. `Ctrl-C` does not raise `SIGINT` in raw mode (`isig` is off), so the
client reads `0x03` itself.
