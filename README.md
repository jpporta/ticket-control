# ticket-control

A personal note-taking server with an analog twist: every task, list, and link you save to the database is also printed as a physical ticket on a thermal receipt printer over TCP.

It's a self-hosted productivity system where each note becomes a piece of paper you can hold, stick on a wall, or hand to someone else. Tasks beep when completed, links print with a scannable QR code, and the day ends with a printed summary.

## What it does

- **Tasks** — create a task via HTTP, get a printed ticket. Mark it done and the printer beeps to acknowledge.
- **Lists** — checklists rendered as a printable receipt.
- **Links** — saves a URL and prints a ticket with a QR code so you can shoot URLs from desktop to phone via paper.
- **Schedules** — cron-driven recurring tasks, with extra predicates like "last weekday of the month" for billing-style chores.
- **End-of-day** — prints a styled summary of what was created and completed.
- **Calendar import** — bulk-import a day's calendar events as tasks (Apple Shortcuts–compatible endpoint).
- **Quiet hours** — printer auto-disables at night; jobs queue up and flush when it turns back on in the morning.

## Tech stack

- **Go 1.24** — `net/http` with the 1.22+ method-aware ServeMux, served over HTTP/2 cleartext (h2c)
- **PostgreSQL** with [`pgx`](https://github.com/jackc/pgx), [`sqlc`](https://sqlc.dev) for type-safe queries and [`goose`](https://github.com/pressly/goose) for migrations
- **ESC/POS** thermal printer over a raw TCP socket via [`hennedo/escpos`](https://github.com/hennedo/escpos)
- **[Typst](https://typst.app)** for receipt layout — templates are rendered to PNG and printed as a bitmap, giving real typographic control (tables, patterns, custom fonts) on an 80mm thermal printer
- **[`robfig/cron`](https://github.com/robfig/cron)** for scheduling
- **Berkeley Mono Nerd Font** for the printed output
- Docker (multi-stage Alpine build) for deployment

## Project layout

```
cmd/
  web/           HTTP server, handlers, middleware, scheduler
  cli/           CLI for creating users and testing the printer
internal/
  application.go business logic entry points (tasks, lists, links, schedules)
  printer/       ESC/POS integration + Typst templates (embedded via go:embed)
  repository/    sqlc-generated database access
  utils/         dithering algorithms, time helpers
migrations/      goose SQL migrations
queries/         sqlc input SQL
requests/        .http files for manual API testing
```

## Notable design decisions

- **Typst → PNG → printer.** Rather than hand-crafting ESC/POS layouts, receipts are typeset in Typst, rasterized, and sent to the printer as an image. This made it trivial to add things like tiled background patterns and proper typography.
- **Print-then-store with rollback.** If a print job fails after the DB insert, the row is deleted so the database and the paper stay in sync.
- **Disabled-printer queue.** When the printer is off (manually or during quiet hours) jobs are kept in memory and flushed when it comes back online.
- **Audible feedback.** Marking a task done sends `ESC i` to make the printer beep — small, satisfying acknowledgement.
- **Daily quotas** (50 tasks, 50 links, 10 lists per user) to prevent runaway paper usage.

## Running it

You need PostgreSQL, the `typst` binary on `PATH`, and a network-attached ESC/POS thermal printer.

```sh
# environment
export DB_URL=postgres://user:pass@host/ticket_control

# database
make up                # apply migrations
make cli name="Your Name"   # create a user, prints an API key

# server
make run               # listens on :8000
```

All endpoints (except `/health`) require an `X-Api-Key` header.

### Docker

```sh
docker build -t ticket-control .
docker run -e DB_URL=... -p 8000:8080 ticket-control
```

The image bundles `typst` and the Berkeley Mono fonts.

## Status

This is a personal project built primarily to learn — to play with `sqlc`, Typst, ESC/POS, h2c, and the Go 1.22 routing improvements while building something physically tangible. It works, it runs, and it prints, but it's not actively maintained as a product.
