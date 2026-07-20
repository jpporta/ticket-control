## Context

The service today mixes concerns: `internal/application.go` exposes a god struct (`Q`, `Cron`, `Printer`) to handlers, which in turn reach across package boundaries to call repository methods and printer methods directly. Domain code in `internal/task.go` and friends has business logic (`TASK_LIMIT`, `print-then-store` rollback), SQL access (`Application.Q.*`), and HTTP-coupled response shaping all in one file. The CLI in `cmd/cli/main.go` is a stub that only prints `time.Now()`; the real helper functions are dead. Two handlers in `cmd/web/endOfDay.go` are never wired to the mux. The printer package owns TCP, ESC/POS, Typst shell-out, and queueing, and services reach into all four via the concrete `*printer.Printer`. Time is read ad-hoc from `time.Now()` in quota checks and cron. Temp `.typ` and `.png` files leak because they're never removed.

The repository uses sqlc-generated code that is the right idea but is hand-imported from handlers via the `Application` struct. The printer is loaded from the `config` table by `printer.New`, but other config values (quiet hours, quotas) are hardcoded constants. There is no place to put a new "policy" without either editing the god struct or scattering constants.

The change is purely structural plus targeted bug fixes. No new business capabilities, no new endpoints, no new schema. The point is to make the next feature cheap to add and to delete the documented landmines.

## Goals / Non-Goals

**Goals:**
- Layered layout that anyone can navigate in 30 seconds.
- A `Printer` port interface so tests and services depend on behaviour, not on a concrete type that owns TCP + Typst + queueing.
- Consistent error model with sentinel codes mapped to HTTP status centrally.
- One place to read time from (`internal/clock`).
- Every printer temp file cleaned up; every error wrapped with `%w`; every `log.Println` replaced with `slog`.
- The 10 documented landmines in `AGENTS.md §12` either fixed or explicitly noted as deferred with reason.
- Bump dependencies within current majors so we pick up bugfixes.

**Non-Goals:**
- New endpoints, new capabilities, schema changes.
- Adding tests beyond what the move forces (we're not writing a new test suite).
- Replacing `hennedo/escpos`, the Typst shell-out, or the pgx driver.
- Adding rate limiting, graceful shutdown, retries, metrics, tracing.
- A pluggable auth backend.

## Decisions

### D1. Layout: domain folders, not flat `internal/*.go`
**Choice.** Move each domain into `internal/<domain>/`. Each domain folder owns its service type, request/response structs, and a tiny `repository` interface that the service depends on (the sqlc-generated `*repository.Queries` satisfies it via a thin adapter in `internal/repository/adapter/`).

**Why.** A `grep` for "task" today hits `cmd/web/handlers.go`, `internal/task.go`, `internal/printer/task.go`, `queries/task.sql`, `repository/task.sql.go` — five files in three packages. With domain folders, all task-related code lives under `internal/task/`. The handler in `cmd/web/handlers.go` only knows the service interface, not the sqlc type.

**Alternatives considered.**
- Keep flat `internal/*.go`, just add interfaces: rejected — packages still grow into grab-bags as new domains are added.
- Full DDD with aggregates/value objects: rejected — overkill for a single-user CRUD service.

### D2. `Printer` is an interface, not a struct
**Choice.** Define `type Printer interface { PrintTask(...); PrintList(...); ...; Toggle(bool); QueueLen() int }` in `internal/service` (or a top-level `internal/ports` package). The concrete implementation stays in `internal/printer`, exported as `*printer.Printer` and constructed via `printer.New(ctx, deps)`.

**Why.** Today every domain file imports `internal/printer` and reaches for `a.Printer.PrintTask`. Tests of `service.CreateTask` either need a real printer or a hand-rolled mock that duplicates the interface. A single interface cuts both: services depend on the interface, tests provide a stub, and the concrete printer is one wiring step in `main`.

**Alternatives considered.**
- Pass `*printer.Printer` everywhere: rejected — every test of business logic would need to mock the whole printer.
- Define interfaces per use-case (`TaskPrinter`, `ListPrinter`, ...): rejected — the queue + disabled semantics belong together; splitting them leaks state.

### D3. Error model: sentinel + wrapping, handler maps to HTTP
**Choice.** Add `internal/apperr` with `var ErrQuotaExceeded = errors.New("quota exceeded")`, etc. Services return `fmt.Errorf("create task: %w", apperr.ErrQuotaExceeded)`. A single helper in `cmd/web` (or `internal/httperr`) walks the chain with `errors.Is` and writes the status code + JSON body.

**Why.** Handlers today do `if err.Error() == "quota exceeded"` style matching in a few places; the rest just return 500 with `err.Error()` in the body, leaking internals. A small table fixes both.

**Alternatives considered.**
- A typed error struct with `Code`, `HTTPStatus`, `Message`: rejected — sentinel + `errors.Is` is what the stdlib ecosystem expects; richer structs add ceremony for no win.
- Per-handler `if err != nil` ladders: rejected — exactly the duplication this design is removing.

### D4. Time goes through one package
**Choice.** `internal/clock` exports `Now() time.Time` (returns `time.Now().UTC()`) and `Today() time.Date` (the UTC date). All quota checks, cron predicates, and "already printed today" checks call `clock.Now()`. The Dockerfile's `TZ=America/Sao_Paulo` remains a presentation hint.

**Why.** Quota windows and "today" must be a single boundary; mixing `time.Now()` and `time.UTC()` at call sites is the kind of bug that only shows up at midnight.

**Alternatives considered.**
- Inject `func() time.Time`: rejected — not worth the test surface for one package.
- Force timezone-aware `time.Time`: rejected — invasive; UTC is fine for a personal service.

### D5. Temp-file hygiene via a `render` helper
**Choice.** `internal/printer/render` exports `Render(tmpl *template.Template, data any) (io.ReadCloser, error)` that creates a tempdir, executes the template to a `.typ` file, shells out to `typst compile -f png`, returns the PNG file (caller `Close()`s), and registers cleanup of the tempdir with a `defer` inside the helper. All `Print<X>` methods go through this helper.

**Why.** Current code leaks both `.typ` and `.png`. The leak is real (each print leaves two files in `$TMPDIR`). Centralising the lifecycle makes it impossible to add a new `Print<X>` without cleanup.

**Alternatives considered.**
- Per-method `defer os.Remove`: rejected — easy to forget on the next method added; same bug as today.
- Stream the PNG to the printer and never touch disk: rejected — typst is a subprocess; we can't avoid the on-disk handoff without a bigger redesign.

### D6. CLI becomes a real dispatcher
**Choice.** `cmd/cli/main.go` uses `flag.NewFlagSet` for subcommands: `user create --name`, `printer test`, `print task`, `print list`, `print image`. The existing helper functions become the bodies. `make cli name="X"` translates to `go run ./cmd/cli user create --name "X"`.

**Why.** The Makefile target and the helper functions already exist; only the dispatch is missing. This is the smallest possible fix to the dead CLI.

**Alternatives considered.**
- Drop the CLI entirely: rejected — `make cli` is in the README workflow and the helper functions are useful for one-off printer tests.

### D7. Dead-code policy
**Choice.** Delete:
- `internal/utils/dither.go`, `internal/utils/rasterize.go` (only referenced by the dead CLI test functions; production printing uses escpos's own dithering).
- `internal/printer/bip.go` and `bip_test.go` only if `printImageTest` is the only consumer after the CLI is fixed. If anything else uses `Bip`, keep it.

Defer (note in tasks):
- `endOfWeekend` and `endOfDayWithTasks` in `cmd/web/endOfDay.go`: either wire them or delete. Decision logged as a task.

**Why.** Dead code is debt that pays no interest; deleting it now is cheaper than deleting it in three feature changes.

### D8. Dependency bump: latest within current majors
**Choice.** `pgx/v5`, `pgx/v5/pgxpool`, `robfig/cron/v3`, `golang.org/x/net`, `hennedo/escpos` → latest within their current major. Run `go mod tidy`. No major bumps.

**Why.** Picks up security and bug fixes without breaking changes. Major bumps (e.g. escpos rewrite, pgx v6) are out of scope.

### D9. Logging: `log/slog`
**Choice.** `slog.New(slog.NewJSONHandler(os.Stdout, ...))` set as default in `main`. Replace `log.Println` calls in handlers and printer with `slog.Info` / `slog.Error` carrying `slog.String("path", r.URL.Path)` etc.

**Why.** Stdlib, no new dep, structured output for free. Log level can be controlled with `LOG_LEVEL=debug` env var.

### D10. Modified-capability decisions
- **tasks / `GetOpenTasks`**: filter by `created_by` in `queries/task.sql::GetOpenTasks` and accept the user id parameter; the handler already passes it through. Regenerate via `make generate`.
- **lists / `DeleteLastList`**: change the inner `SELECT` from `task` to `list`; regenerate. No service signature change — the function already takes a list id.

## Risks / Trade-offs

- **Big diff, small behaviour change.** → Land in one PR with a smoke-test section in the tasks ("curl every endpoint after the move"). If anything breaks, the regression is easy to bisect.
- **Refactor breaks the cron job's in-memory `jobs` map.** The scheduler stores `cron.EntryID` per `schedule.id`; if we touch the `Application` type, the in-memory state could reset on restart in a way that's visible. → Tasks keep `CronJob` as a single instance held in the service container; no change to its lifecycle.
- **`go mod tidy` may surface transitive bumps that don't match the spirit of "current majors".** → Lock the four direct deps to their current-major latest before running tidy; pin with `require ... // indirect` only if necessary.
- **Typst font not in repo** (existing AGENTS.md issue). → Out of scope; documented as a precondition for `make run`. We do not add a fixture font.
- **The two unwired handlers in `endOfDay.go`** — wiring them changes behaviour; deleting them is a small loss of intent. → Tasks list both options; default is to wire `endOfWeekend` only if it has a clear caller, else delete both.
- **Race between cron startup and config load.** The printer is constructed in `printer.New` which reads the `config` table; if the DB is unreachable at startup, current code panics. The refactor preserves this behaviour (don't fix what's not broken here).
- **Tests we break.** `internal/printer/printer_test.go`, `bip_test.go`, and `internal/utils/time_test.go` will need import-path updates. → Tasks include a `go test ./...` step after the move.
- **sqlc regenerated code may move functions if query text changes.** We only edit `task.sql::GetOpenTasks` and `list.sql::DeleteLastList`; the regen diff stays small.

## Migration Plan

Single-environment, single-binary service. No data migration. No schema migration.

1. Land the refactor behind the existing routes — no API surface change.
2. `make generate` after editing `queries/task.sql` and `queries/list.sql`.
3. `go mod tidy` + `go build ./...` + `go test ./...`.
4. Smoke-test: walk every endpoint in `requests/*.http` with `curl` against `localhost:8000` after `make run`.
5. Rollback: revert the single commit. No DB state to restore, no schema to revert.

## Open Questions

- Should `endOfWeekend` and `endOfDayWithTasks` be wired or deleted? (Decided in tasks: default to delete unless a caller is named.)
- Should we keep the `Bip`/`TestBip` printer test path at all, or move it to a `_test.go` example? (Decided in tasks: delete if no consumer.)
- Does the printer's `Queue` belong on the interface or stay internal? (Decided: stays internal — callers only need `Print<X>` and `Toggle`.)
