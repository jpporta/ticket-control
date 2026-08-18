# Typewriter mode

## Why

Making a quick list today means composing a JSON body and POSTing it, then waiting
for the Typst → PNG → raster round trip. For throwaway lists and notes that is more
ceremony than the artifact deserves.

Typewriter mode is a terminal client that turns the thermal printer into a literal
typewriter: start it, type, each line commits to paper as you press Enter, cut when
you're done. Nothing is stored, nothing is rendered, nothing is queued.

## What

A new `typewriter` subcommand on the existing CLI that talks **directly** to the
printer over TCP :9100. No HTTP endpoint, no WebSocket, no database, no persistence.

The client holds the current line so you can backspace; Enter commits it over a
short-lived TCP connection. Keys: `Enter` commit · `Esc` cut and continue ·
`Ctrl-D` cut and exit · `Ctrl-C` abort without cutting.

When stdin is not a terminal the same subcommand prints the piped input as one
ticket and cuts, so `cal | ticket-control typewriter` works.

Also: `printer.start()` gains a short dial retry, so a server-side print that
collides with a typewriter session recovers instead of rolling back its DB row.

## Non-goals

- Persistence of any kind. Typed text exists on paper only.
- Quotas, auth, user accounting. The client never touches the server or the DB.
- Typst rendering. Typewriter output uses the printer's built-in font.
- A per-keystroke transmission mode. It was built and tested against the
  hardware: the printer commits on `LF` regardless, so the paper output was
  identical to buffered mode and the only difference was losing backspace.
  Removed.
