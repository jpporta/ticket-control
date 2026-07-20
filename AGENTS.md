# AGENTS.md — working notes for ticket-control

This file is for AI coding agents (and humans) picking up the project. It documents
what's actually here, how it fits together, how to extend it, and where the landmines
are. The project's README.md is the high-level pitch; this is the field guide.

---

## 1. What this project is

A Go HTTP service that turns API calls into **physical printed tickets** on an 80mm
ESC/POS thermal printer (raw TCP, port 9100). Every task, list, link, end-of-day
summary, and cron-generated task is both persisted in Postgres **and** printed.

Receipts are typeset with **Typst** → PNG → sent to the printer as a raster image.
This gives real typography (tables, patterns, custom fonts) on a thermal printer
that natively only does line-text and bitmaps.

Personal project. Not production. Quotas, a single printer, a single Postgres,
an X-Api-Key auth scheme. Don't over-engineer.

---

## 2. Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.24.5 |
| HTTP | stdlib `net/http` with method-aware `ServeMux` (Go 1.22+), served over h2c |
| DB driver | `github.com/jackc/pgx/v5` |
| Queries | `sqlc` (typed, generated; **do not edit generated files**) |
| Migrations | `pressly/goose` |
| Printer | `github.com/hennedo/escpos` over `net.Dial("tcp", ...)` |
| Scheduler | `github.com/robfig/cron/v3` |
| Receipt layout | `typst` CLI invoked via `os/exec` |
| Font | BerkeleyMono Nerd Font (font files are gitignored — must be supplied under `static/fonts/`) |
| Deployment | Multi-stage Alpine Dockerfile |

---

## 3. Layout

Layered: `cmd` → `service` → `<domain>` → `repository` → `printer`. Each layer
depends only on the layer immediately below.

```
cmd/
  web/             HTTP server entrypoint, handlers, middleware, slog init
    main.go        wires the *pgxpool.Pool, *printer.Printer, and *service.Services
    handlers.go    thin HTTP handlers that depend only on *service.Services
    middleware.go  logRequestMiddleware + authMiddleware
    httperr.go     central error → HTTP status mapper (apperr.Kind → status)
    endOfDay.go    end-of-day / weekend handler
    schedule.go    schedule handlers
    calendar.go    POST /events (Apple Shortcuts envelope)
  cli/             subcommand dispatcher (user/printer/print)
internal/
  clock/           Now() and Today() — single source of wall-clock time
  apperr/          sentinel errors (ErrQuotaExceeded, ErrPrinterOffline, etc.)
  ports/           Printer interface (the service-facing port)
  service/         service container; constructs every domain Service
  task/            task domain (service.go, types.go, quota.go)
  list/            list domain
  link/            link domain
  schedule/        schedule domain + cron scheduler
  endofday/        end-of-day / weekend summaries
  events/          calendar-import (Apple Shortcuts)
  printer/         ESC/POS + Typst integration
    *.go           one file per printable artifact (task, list, link, end_of_day, text, bip)
    render/        temp-file-lifecycle helper (executes Typst, cleans up both files)
    models/*.typ   Typst templates, embedded via go:embed
  repository/      sqlc-GENERATED. Do not edit. Regenerate via `make generate`.
    adapter/       per-domain adapters that satisfy each domain's Repo interface
  utils/           date predicates, random key
migrations/        goose migrations, one file per change
queries/           sqlc input (the source of truth for SQL)
requests/          .http files for manual API testing (use with the REST Client extension)
openspec/          OpenSpec specs and changes (see §10)
.opencode/         opencode CLI config (skills, commands)
```

---

## 4. The print flow (read this before touching printer code)

This is the single most important flow in the project. **Every** printed artifact goes through it.

```
HTTP request
   │
   ▼
cmd/web/handlers.go (parse JSON, pull userId from context)
   │
   ▼
internal/service (service.Printer is the ports.Printer interface)
   │
   ▼
internal/<domain>/service.go (e.g. task.Service.CreateTask)
   │
   ├── 1. INSERT row into Postgres via internal/repository (sqlc, through adapter)
   │       if it fails: return apperr-mapped error. Nothing has been printed yet.
   │
   ├── 2. Call s.printer.Print<X>(...)  (the ports.Printer interface)
   │       if it fails: DELETE the row you just inserted, return the wrapped error.
   │       (This is "print-then-store with rollback" — see internal/task/service.go,
   │        internal/list/service.go, internal/link/service.go.)
   │
   ▼
internal/printer/<x>.go (e.g. task.go)
   │
   ├── 3. Check p.Enabled. If false, append a closure to p.queue and return errPrinterOffline.
   │       (The queue flushes when p.Toggle(true) is called.)
   │
   ├── 4. Open a TCP socket to p.IP:p.Port. Set p.e = escpos.New(socket).
   │       defer close().  (See printer.go start().)
   │
   ├── 5. Call render.Render(tmpl, data, name) — see internal/printer/render/render.go.
   │       Executes the template, shells out to typst, returns the PNG ReadCloser +
   │       a cleanup func that removes both .typ and .png.
   │
   ├── 6. image.Decode the PNG, then render.CropHeight8 (height % 8 == 0).
   │
   ├── 7. p.Reset() (ESC @, ESC R 0) → p.printImage(img) (PrintImage + PrintAndCut).
   │
   └── 8. defer cleanup() and defer close() run → temp files gone, socket closed, p.e = nil.
```

### Gotchas specific to this flow

- **Temp files are cleaned up.** The render helper owns the lifecycle. Don't bypass it.
- **Typst template parsing happens once at startup** (`loadTemplates` in text.go). Adding
  a new template means: drop the .typ file under `internal/printer/models/`, add the name
  to the loop in `loadTemplates`, then write a `Print<X>` method.
- **8-pixel crop.** ESC/POS raster bitmaps need height % 8 == 0. Use `render.CropHeight8`.
- **The escpos library is small and quirky.** It uses `e.WriteRaw([]byte{...})` for raw
  bytes and `e.PrintImage` / `e.QRCode` / `e.PrintAndCut` for the higher-level ops.
  See `link.go` for an example of stacking an image + a QR code + a cut.
- **Queue-on-disabled is closure-based.** Each `Print<X>` enqueues `func() error { return p.Print<X>(...) }`.
  When the printer is re-enabled, the queue drains with a 1-second sleep between jobs.
  Don't store state inside the closure other than the inputs — it captures by reference.
- **`Printer` is an interface.** Defined in `internal/ports`; `*printer.Printer` satisfies
  it via a compile-time assertion. Tests can supply stubs without importing TCP/Typst.

---

## 5. Adding a new capability

Step-by-step for, say, a new "note" artifact:

1. **Migration.** `make new_migration name=notes` → edit the generated `.sql` →
   `make up`.
2. **SQL.** Add `queries/note.sql` with `-- name: CreateNote :one`, etc.
   Run `make generate`. New `internal/repository/note.sql.go` appears.
3. **Business logic.** New `internal/note.go` with an `Application.CreateNote` method.
   Follow the print-then-store-with-rollback pattern (§4).
4. **Domain JSON struct** (request body) in `cmd/web/handlers.go` or a sibling file.
5. **Handler** in `cmd/web/handlers.go` or a new file under `cmd/web/`. Register
   the route in `cmd/web/main.go` `mux.HandleFunc("METHOD /path", protectedRoute(h.handler))`.
6. **Print method** in `internal/printer/note.go` (copy `task.go`). Add a Typst
   template under `internal/printer/models/note.typ` and wire it in `loadTemplates`.
7. **Test request.** Add `requests/createNote.http` with an `X-Api-Key: {{API_KEY}}`
   placeholder.
8. **OpenSpec change** (see §10).

Always end with `make run` and a manual curl against `localhost:8000`.

---

## 6. Daily quotas

Hardcoded in `internal/{task,list,link}/`:

- `TaskLimit = 50` per user per UTC day
- `LinkLimit = 50`
- `ListLimit = 10`

Quota check happens in the service **before** the DB insert. If you add a new
artifact type, add a `<X>LimitReached` check in the same shape.

Note: the day boundary is UTC (`clock.Today()`). If you want local-time quotas,
change the `clock.Now()` derivation in the limit functions.

## 7. Scheduler

`internal/schedule/service.go`:

- Two **internal jobs** are registered on `Start` (always on):
  - `0 22 * * *` → `printer.Toggle(false)` (quiet hours start)
  - `0 8 * * *` → `printer.Toggle(true)` (quiet hours end)
- **User-defined jobs** come from the `schedule` table where `enabled = TRUE`.
  Each row has a cron expression and an optional `check_function` that gates
  execution (e.g. "last weekday of the month" for billing chores).

  Available check_functions (defined as `schedule.CheckFunc` values, implemented
  in `internal/utils/time.go`):
  - `is_last_workday_of_month`
  - `is_last_weekday_of_middle` (last workday on/before the 15th)
  - `is_last_weekday_of_10` (last workday on/before the 10th)

  Adding a new predicate: write `func Is<Name>(fn func())` in `internal/utils/time.go`,
  and register it in the map passed to `schedule.New` in `cmd/web/main.go`.

When a scheduled job fires, it calls the `taskCallback` configured at `schedule.New`
with priority 0 (which maps to the "no priority" icon in the Typst template).

**Schedule state** lives in two places: the `schedule` table AND the in-memory
`Service.jobs map[int32]cron.EntryID`. Toggling in the DB doesn't take effect until
the server restarts — toggling via `PUT /schedule?id=N` keeps both in sync.

---

## 8. Auth

All endpoints except `/health` require `X-Api-Key: <key>`.

Flow in `cmd/web/middleware.go`:

1. `logRequestMiddleware`:
   - Reject if header missing → 401 (apperr.ErrUnauthorized).
   - Look up via `adapter.User.GetUserByKey` → user ID + name.
   - Log the access via `adapter.Access.AddAccess`.
   - Stash user ID and name in `r.Context()` under typed keys.
2. `authMiddleware`:
   - Returns 401 if `userId == 0` (i.e. lookup failed but we let the request through
     the previous step anyway).

There's no rate limiting, no expiry, no revocation flow. Users are created via
`make cli name="..."` which dispatches to `go run ./cmd/cli user create --name "..."`.
The CLI uses the sqlc-generated `CreateUser` directly (no domain service for users yet).

## 9. Apple Shortcuts compatibility

`POST /events` takes the body `{"body": "<JSON string>"}` — a nested JSON envelope
because Apple Shortcuts doesn't let you set a raw request body directly. The handler
in `cmd/web/calendar.go` unmarshals the outer object, then unmarshals `body` again.
**Do not "fix" this** — it's a public contract with shortcuts the user has wired up.

## 11. Build / run / test

```sh
# install tooling (one-time)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest

# environment
cp .env.example .env   # then edit DB_URL, PRINTER_IP, PRINTER_PORT

# schema
make up                # apply goose migrations

# regenerate sqlc if you touched queries/
make generate

# run the server (loads .env via the Makefile)
make run               # listens on :8000 over h2c

# create a user
make cli name="Your Name"

# watch a Typst template during edits
make typst-task
```

Tests today: only `internal/utils/time_test.go` (pure functions, safe to run anywhere).
`go test ./...` runs in any environment.

Docker:
```sh
docker build -t ticket-control .
docker run --env-file .env -p 8000:8080 ticket-control
```
The image bundles `typst` and the fonts from `static/fonts/`. Since fonts are
gitignored, you must mount or bake them into the build context yourself.

---

## 12. Known issues (do NOT fix without an explicit change proposal)

These are landmines an agent will trip over if it doesn't know. Documented here so
they don't get re-discovered the hard way.

1. **Font not in repo.** `static/fonts/*` is gitignored. The Typst templates
   hardcode `font: "BerkeleyMono Nerd Font"`. The Docker build copies from
   `static/fonts/`, so without the font file locally, both `docker build` and
   `make run` will fail at print time with a font-related Typst error. The
   project README mentions this is a paid font — supply it yourself.

2. **`internal/events.Service.UserPrintedSinceUTCMidnight`** uses `noPrintedToday > 1`
   (off-by-one). Reads the `access` table it itself wrote to in `logRequestMiddleware`.
   Fragile; documented in code.

3. **No graceful shutdown.** `cmd/web/main.go` traps signals but `os.Exit(1)`s
   hard. The cron job is stopped, but in-flight requests and queued print jobs
   are dropped. If that matters, add `server.Shutdown(ctx)` to the signal path.

4. **No timezone awareness.** `TZ=America/Sao_Paulo` in the Dockerfile but the
   app uses `clock.Now()` (UTC) everywhere except the final string formatting.
   Quota windows, end-of-day offsets, and "today" all assume UTC day boundary.

5. **`printImage` direct call in `link.go`.** Most `Print<X>` routes go through
   the internal `printImage` wrapper (`printer.go` → `image.go`) which does
   `PrintImage` + `PrintAndCut`. `link.go` calls `p.e.PrintImage` directly and
   then explicitly calls `p.e.PrintAndCut`. If you change the wrapper, audit
   `link.go` too.

6. **CLI's `print image` subcommand is gone.** The original helper relied on
   the now-deleted `internal/utils/dither.go`. Re-add when there's a real
   image-printing feature to support.

7. **`errPrinterOffline` is re-declared per-domain.** Each of `task`, `list`,
   `link`, `endofday` has its own `errors.New("printer offline")` because the
   `internal/printer` package keeps it unexported. If you ever want a single
   source of truth, export it from the printer package.

Items resolved by the refactor (kept here for reference):
- ~~`queries/list.sql::DeleteLastList` selects from `task` not `list`~~ → fixed.
- ~~`cmd/web/endOfDay.go` dead handlers~~ → deleted.
- ~~`cmd/cli/main.go` stub~~ → rewritten as a real subcommand dispatcher.
- ~~`internal/utils/dither.go` + `rasterize.go`~~ → deleted.
- ~~`getOpenTasks` ignores `created_by`~~ → fixed in `queries/task.sql`.
- ~~Temp `.typ`/`.png` file leaks~~ → fixed via `internal/printer/render`.
- ~~Mixed `time.Now()` / `time.UTC()`~~ → centralised in `internal/clock`.
- ~~Ad-hoc `fmt.Errorf` and `http.Error` leaking internals~~ → fixed via
  `internal/apperr` + `cmd/web/httperr.go`.
- ~~`log.Println` in handler and printer code~~ → replaced with `log/slog`.

---

## 13. Quick reference — where to look

| You want to... | Look here |
|---|---|
| Add a new HTTP endpoint | `cmd/web/main.go` (mux), `cmd/web/handlers.go` (handler), `internal/<domain>.go` (logic) |
| Add a new SQL query | `queries/<x>.sql`, then `make generate` |
| Change the schema | new file in `migrations/`, `make up` |
| Change how a ticket looks | `internal/printer/models/<x>.typ`, then rebuild |
| Add a new printable artifact | `internal/printer/<x>.go` + `models/<x>.typ` + register in `loadTemplates` |
| Add a cron predicate | `internal/utils/time.go` + register in `internal/cron.go funcMap` |
| Change the printer's host/port | DB row `config.key='thermal-printer'` (`UPDATE config SET value = '{"ip":"...","port":...,"enabled":true}' WHERE key='thermal-printer';`) — or use the `.env` if you wire it through |
| Change daily quotas | `internal/task.go` (`TASK_LIMIT`), `internal/list.go` (`LIST_LIMIT`), `internal/link.go` (`LINK_LIMIT`) |
| Change quiet hours | `internal/cron.go createInternalJobs` |
| Debug a print job | check the `internal/printer/<x>.go` flow (§4), then `internal/utils/dither.go` if you suspect dithering, then `internal/utils/rasterize.go` |
