## Why

The codebase has grown organically and the seams show: a CLI stub that only prints `time.Now()`, two dead handlers in `endOfDay.go`, a `DeleteLastList` query that targets the wrong table, `getOpenTasks` that ignores the user filter, printer temp files that leak, mixed UTC/`time.Now()` time handling, and no separation between HTTP wiring and business logic. None of these are functional showstoppers, but they are the tax we pay every time we add a feature. This change lays down a clean-architecture skeleton (transport / service / repository / printer as separate layers with explicit interfaces) and uses the move to fix every documented landmine in `AGENTS.md §12` in one pass — leaving the running behaviour unchanged.

## What Changes

- **Layout reshape.** Move domain logic out of `internal/*.go` (which mixes types, queries, and orchestration) into `internal/<domain>/` packages (one folder per bounded context: `task`, `list`, `link`, `schedule`, `endofday`, `events`). Keep `internal/repository/` as the sqlc-generated data layer. Add an explicit `internal/service/` package that holds use-cases and depends only on repository interfaces and a `Printer` port.
- **Hexagonal port for the printer.** `internal/printer` exposes a `Printer` interface (the only thing services touch); the TCP/ESC-POS/Typst implementation moves to `internal/printer/escpos/`. The `Queue` and quiet-hours machinery stay on the interface, not the concrete type.
- **DI via constructor.** `cmd/web/main.go` builds the concrete `*pgxpool.Pool`, the concrete printer, wires them into a service container, and hands the container to thin handlers. No more package-level globals; no more `Application` god-struct that exposes `Q`, `Cron`, `Printer` directly to handlers.
- **Errors with context.** Replace ad-hoc `fmt.Errorf`/string returns with `errors.Join` and `%w` everywhere; introduce a small `apperr` package with sentinel codes (`ErrQuotaExceeded`, `ErrPrinterOffline`, `ErrNotFound`) that handlers map to HTTP status codes.
- **Time.** All quota and "today" checks use `time.Now().UTC()` consistently. The Dockerfile `TZ` stays as a presentation hint only; we do not depend on it for correctness.
- **Temp-file hygiene.** Every `os.CreateTemp` is paired with `defer os.Remove(name)`. A small helper in `internal/printer/render` owns the temp-dir lifecycle.
- **Dependency bump (current majors only).** `pgx/v5` → latest v5; `pgx/v5/pgxpool` → latest; `pressly/goose/v3` (tool only); `robfig/cron/v3` → latest; `hennedo/escpos` → latest tagged; `golang.org/x/net` → latest. Run `go mod tidy` and `go build ./...` to verify.
- **Dead code removed.** Delete `cmd/web/endOfDay.go::endOfWeekend` and `::endOfDayWithTasks` (or wire them up — see tasks). Delete `internal/utils/dither.go` + `rasterize.go` (only referenced by the dead CLI test functions). Delete `internal/printer/bip_test.go` if it's the only consumer of `printImageTest` once the CLI is fixed.
- **CLI wired up.** `cmd/cli/main.go` becomes a real subcommand dispatcher (`user create`, `printer test`, `print task`, `print list`, `print image`) using `flag.NewFlagSet`. The existing helpers get called.
- **Bug fixes shipped with the move.** `queries/list.sql::DeleteLastList` selects from `list` not `task`. `internal/task.go::GetOpenTasks` filters by `created_by`. The off-by-one in `events.go::hasUserAlreadyPrintedToday` is documented as intentional and the function is renamed for clarity.
- **Logging.** Introduce `log/slog` (stdlib) with a JSON handler in non-interactive mode, replace ad-hoc `log.Println` in handlers and printer.
- **No behaviour change for callers.** Every existing HTTP route returns the same status codes for the same inputs. The Apple Shortcuts `/events` envelope is preserved. Quiet-hours cron schedule is preserved. Quota numbers are preserved.

Out of scope (explicit non-goals for this change):
- Adding tests beyond what the move forces us to touch (no new test suite).
- Adding rate limiting, retries, graceful shutdown beyond the existing signal trap.
- Multi-tenant auth, OAuth, JWT.
- Replacing the ESC/POS / Typst pipeline.

## Capabilities

### New Capabilities
- `architecture`: documents the layer boundaries (transport → service → repository → printer), dependency direction, and the `Printer` port interface.
- `error-model`: defines `apperr` sentinels and the handler → HTTP mapping table.
- `cli`: documents the actual CLI subcommands and their arguments.
- `printer-port`: documents the `Printer` interface and the queue/disabled semantics that callers depend on.

### Modified Capabilities
- `tasks`: `GetOpenTasks` now requires and applies the `created_by` filter (requirement change, not just implementation).
- `lists`: `DeleteLastList` must target the `list` table (requirement change: the rollback path must delete the row just inserted).

## Impact

- **Files touched.** Most files under `internal/`, `cmd/`, and `queries/list.sql` move or are rewritten. `migrations/` is untouched (no schema change). `internal/repository/` is regenerated only if a query string changes.
- **APIs.** None of the HTTP endpoints change. `/health` still answers unauthenticated. `X-Api-Key` still required for everything else. The `/events` body envelope is unchanged.
- **Build.** `go.mod` bumps within current majors; `go.sum` rotates. `make generate` is rerun. `make build` and `make run` are the verification commands.
- **Runtime.** No new processes, no new env vars. The same single Postgres + one printer + one h2c `:8000` listener.
- **Risk.** High churn in source layout, low behavioural risk because every handler is kept thin and the service-layer signatures are stable. The biggest correctness risk is the `DeleteLastList` fix (only triggers on a print failure after list insert); mitigated by the existing manual end-to-end test flow in `requests/*.http`.
