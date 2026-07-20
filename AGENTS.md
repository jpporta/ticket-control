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

```
cmd/
  web/             HTTP server entrypoint, handlers, middleware, scheduler bootstrap
  cli/             CLI entrypoint (see §12 — currently a no-op stub)
internal/
  application.go   service object: Q (sqlc), Cron, Printer
  task.go          task domain logic
  list.go          list domain logic
  link.go          link domain logic
  schedule.go      schedule domain logic
  end_of_day.go    end-of-day / weekend summaries
  events.go        calendar-import handler logic
  cron.go          cron scheduler + custom check-functions
  printer/         ESC/POS + Typst integration
    *.go           one file per printable artifact (task, list, link, end_of_day, text, bip, image)
    models/*.typ   Typst templates, embedded via go:embed
  repository/      sqlc-GENERATED. Do not edit. Regenerate via `make generate`.
  utils/           date predicates, random key, dithering (mostly dead — see §12)
migrations/        goose migrations, one file per change
queries/           sqlc input (the source of truth for SQL)
requests/          .http files for manual API testing (use with the REST Client extension)
openspec/          OpenSpec specs and changes (see §10)
.opencode/         opencode CLI config (skills, commands)
```

---

## 4. The print flow (read this before touching printer code)

This is the single most important flow in the project. **Every** printed artifact goes
through it.

```
HTTP request
   │
   ▼
cmd/web/handlers.go (parse JSON, pull userId from context, check quota)
   │
   ▼
internal/<domain>.go (Application method)
   │
   ├── 1. INSERT row into Postgres via internal/repository (sqlc)
   │       if it fails: return 500. Nothing has been printed yet.
   │
   ├── 2. Call a.Printer.Print<X>(...)
   │       if it fails: DELETE the row you just inserted, return 500.
   │       (This is "print-then-store with rollback" — see internal/task.go:32-54,
   │        internal/list.go:31-59, internal/link.go:31-58.)
   │
   ▼
internal/printer/<x>.go (e.g. task.go)
   │
   ├── 3. Check p.Enabled. If false, append a closure to p.queue and return an error.
   │       (The queue flushes when TooglePrinter(true) is called — see cron.go:65.)
   │
   ├── 4. Open a TCP socket to p.IP:p.Port. Set p.e = escpos.New(socket).
   │       defer close().  (See printer.go start().)
   │
   ├── 5. Execute a Go text/template over the embedded .typ file → write to os.CreateTemp("", "x-*.typ")
   │
   ├── 6. exec.Command("typst", "c", file.Name(), "-f", "png").Run()
   │       typst must be on PATH. The font must be discoverable by fontconfig
   │       (the Dockerfile runs fc-cache after copying fonts into /usr/share/fonts/truetype/).
   │
   ├── 7. Open <file>.png, image.Decode, crop to multiple of 8 px tall (printer constraint).
   │
   ├── 8. p.Reset() (ESC @, ESC R 0) → p.printImage(img) (PrintImage + PrintAndCut)
   │
   └── 9. defer close() runs → socket.Close(), p.e = nil.
```

### Gotchas specific to this flow

- **Temp files leak.** The PNG is read but never explicitly removed. The .typ is
  never removed. Don't copy that pattern when adding new types — defer `os.Remove(file.Name())`.
- **Typst template parsing happens once at startup** (`loadTemplates` in text.go). Adding
  a new template means: drop the .typ file under `internal/printer/models/`, add a
  `template.New("name").Parse(...)` call in `loadTemplates`, then write a `Print<X>`
  method.
- **8-pixel crop.** ESC/POS raster bitmaps need height % 8 == 0. Every Print<X>
  method crops the image. If you're writing a new one, copy that crop block.
- **The escpos library is small and quirky.** It uses `e.WriteRaw([]byte{...})` for raw
  bytes and `e.PrintImage` / `e.QRCode` / `e.PrintAndCut` for the higher-level ops.
  See `link.go` for an example of stacking an image + a QR code + a cut.
- **Queue-on-disabled is closure-based.** Each `Print<X>` enqueues `func() error { return p.Print<X>(...) }`.
  When the printer is re-enabled, the queue drains with a 1-second sleep between jobs.
  Don't store state inside the closure other than the inputs — it captures by reference.

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

Hardcoded in `internal/{task,list,link}.go`:

- `TASK_LIMIT = 50` per user per UTC day
- `LINK_LIMIT = 50`
- `LIST_LIMIT = 10`

Quota check happens in the handler **before** the DB insert. If you add a new
artifact type, add a `<X>LimitReached` check in the same shape.

Note: the day boundary is UTC (`time.UTC`). If you want local-time quotas, change
the `time.Now().Date()` derivation in the limit functions.

---

## 7. Scheduler

`internal/cron.go`:

- Two **internal jobs** are registered on `Start` (always on):
  - `0 22 * * *` → `a.Printer.TooglePrinter(false)` (quiet hours start)
  - `0 8 * * *` → `a.Printer.TooglePrinter(true)` (quiet hours end)
- **User-defined jobs** come from the `schedule` table where `enabled = TRUE`.
  Each row has a cron expression and an optional `check_function` that gates
  execution (e.g. "last weekday of the month" for billing chores).

  Available check_functions (defined in `internal/cron.go`, implemented in
  `internal/utils/time.go`):
  - `is_last_workday_of_month`
  - `is_last_weekday_of_middle` (last workday on/before the 15th)
  - `is_last_weekday_of_10` (last workday on/before the 10th)

  Adding a new predicate: write `func Is<Name>(fn func())` in `internal/utils/time.go`,
  add the string to `internal.PossibleCheckFunctions`, and register it in the
  `funcMap` at the top of `internal/cron.go`.

When a scheduled job fires, it calls `a.CreateTask(ctx, title, description, 0, createdBy)`
with priority 0 (which maps to the "no priority" icon in the Typst template).

**Schedule state** lives in two places: the `schedule` table AND the in-memory
`CronJob.jobs map[int32]cron.EntryID`. Toggling in the DB doesn't take effect until
the server restarts — toggling via `PUT /schedule?id=N` keeps both in sync.

---

## 8. Auth

All endpoints except `/health` require `X-Api-Key: <key>`.

Flow in `cmd/web/middleware.go`:

1. `logRequestMiddleware`:
   - Reject if header missing.
   - `GetUserByKey(key)` → user.
   - `AddAccess(user_id, ip, path, method)` for the access log.
   - Stash `userId` and `userName` in `r.Context()`.
2. `authMiddleware`:
   - Returns 401 if `userId == 0` (i.e. lookup failed but we let the request through
     the previous step anyway).

There's no rate limiting, no expiry, no revocation flow. Users are created via
`cmd/cli` (currently broken — see §12). For development you can also create them
with raw SQL: `INSERT INTO public."user"(name, api_key) VALUES ($1, $2);`.

---

## 9. Apple Shortcuts compatibility

`POST /events` takes the body `{"body": "<JSON string>"}` — a nested JSON envelope
because Apple Shortcuts doesn't let you set a raw request body directly. The handler
unmarshals the outer object, then unmarshals `body` again. **Do not "fix" this** —
it's a public contract with shortcuts the user has wired up.

---

## 10. OpenSpec workflow

This repo uses [OpenSpec](https://github.com/...) for spec-driven change management.

```
openspec/
  config.yaml      context block (project conventions — read this)
  specs/           main specs (the current source of truth for capabilities)
  changes/         proposed changes, one folder per change
    archive/       archived (merged) changes
```

Workflow:
- `/opsx-explore` — think about something, no artifacts written.
- `/opsx-propose <name>` — generate a proposal (proposal.md + design.md + tasks.md + delta specs).
- `/opsx-apply <name>` — implement the tasks.
- `/opsx-archive <name>` — fold the delta specs into main specs, move change to archive.

Before writing code, check if a change already exists for your work. If not,
propose one. Read `openspec/config.yaml` for the project context that gets
injected into every artifact.

Currently: no active changes, no main specs. The context in `config.yaml` is
the only artifact. First proposals should probably seed main specs for the
existing capabilities (tasks, lists, links, schedules, end-of-day, events, printer)
so future changes have a baseline.

---

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

# create a user (currently broken — see §12)
make cli name="Your Name"

# test the printer directly
make test-printer      # only runs TestBip in internal/printer/

# watch a Typst template during edits
make typst-task
```

Tests today: only `TestBip` (tries to actually print — needs a real printer) and
`internal/utils/time_test.go` (pure functions, safe to run anywhere).
`make test-printer` will fail without a printer on `os.Getenv("DB_URL")`-reachable
network.

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

1. **`queries/list.sql` `DeleteLastList`** — inner `SELECT` is from `task` instead
   of `list`. Under list-create rollback, this deletes the wrong row. Latent because
   list rollbacks are rare. Fix in `queries/list.sql` then `make generate`.

2. **`cmd/web/endOfDay.go` has two dead handlers.** `endOfWeekend` (line 12) and
   `endOfDayWithTasks` (line 75) are defined but **never registered** on the mux
   in `cmd/web/main.go`. Only `endOfDay` and `endOfDayAuto` are wired. Either
   register them with routes or delete the dead code.

3. **`cmd/cli/main.go` is a stub.** `main()` only prints the current time. The
   real CLI helpers (`createUser`, `createLinkTest`, `printListTest`,
   `printTaskTest`, `printImageTest`, `simpleTest`) are dead — the Makefile runs
   `go run ./cmd/cli/* --name="..."` expecting `createUser` to fire, but it
   doesn't, because nothing calls it. Fix: wire up subcommands (e.g. with
   `flag.NewFlagSet`) or convert to a simple `flag.Parse` + dispatch in `main()`.

4. **`internal/utils/dither.go` + `internal/utils/rasterize.go`** — these
   dithering algorithms are only referenced from the dead `printImageTest`
   function in `cmd/cli/main.go`. Production printing uses escpos's `PrintImage`
   directly, which does its own dithering. If you want to revive image printing,
   keep this code; otherwise it's dead weight.

5. **`internal/events.go` "already printed today"** — `hasUserAlreadyPrintedToday`
   uses `noPrintedToday > 1` (off-by-one). Intentional? Unclear. Reads the
   `access` table it itself wrote to in `logRequestMiddleware`. Fragile.

6. **Font not in repo.** `static/fonts/*` is gitignored. The Typst templates
   hardcode `font: "BerkeleyMono Nerd Font"`. The Docker build copies from
   `static/fonts/`, so without the font file locally, both `docker build` and
   `make run` will fail at print time with a font-related Typst error. The
   project README mentions this is a paid font — supply it yourself.

7. **`getOpenTasks` ignores `created_by`.** Look at `queries/task.sql` line 21-26:
   it selects all open tasks, not just the user's. The handler in
   `cmd/web/handlers.go:getOpenTasks` passes `userId` into `Application.GetOpenTasks`
   but `Application.GetOpenTasks` (`internal/task.go:99`) doesn't pass it to the
   query. Either filter by user (preferred — there's no use case for seeing other
   people's open tasks) or document it as intentional.

8. **Two paths through `p.e.PrintImage`.** Most calls go through the internal
   `printImage` (printer.go → image.go) which does `PrintImage` + `PrintAndCut`.
   `internal/printer/link.go:78` calls `p.e.PrintImage` directly. The cut still
   happens at the end of the link flow, but only because `PrintAndCut` is called
   explicitly on line 88. If you change the wrapper, audit both call sites.

9. **No graceful shutdown.** `cmd/web/main.go` traps signals but `os.Exit(1)`s
   hard. The cron job is stopped, but in-flight requests and queued print jobs
   are dropped. If that matters, add `server.Shutdown(ctx)` to the signal path.

10. **No timezone awareness.** `TZ=America/Sao_Paulo` in the Dockerfile but the
    app uses `time.Now()` and `time.UTC` interchangeably. Quota windows, end-of-day
    offsets, and "today" all assume the server clock's UTC day boundary.

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
