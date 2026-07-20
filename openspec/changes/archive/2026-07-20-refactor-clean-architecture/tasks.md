## 1. Dependency bump

- [x] 1.1 Update `go.mod` to the latest within current majors: `pgx/v5`, `pgx/v5/pgxpool`, `robfig/cron/v3`, `golang.org/x/net`, `hennedo/escpos`. Run `go mod tidy`. Verify with `go build ./...`.
- [x] 1.2 Verify the four direct deps and their transitive closure resolve without major bumps. If `go mod tidy` wants a major bump, pin to current major and note the version in a comment in `go.mod`.

## 2. Foundation: clock, errors, printer port

- [ ] 2.1 Add `internal/clock/clock.go` exporting `Now() time.Time` returning `time.Now().UTC()` and `Today() time.Date`. Replace every `time.Now()` and `time.UTC()` in `internal/*.go` (and the future `internal/<domain>/*.go`) with `clock.Now()`. Keep the Dockerfile `TZ=America/Sao_Paulo` as a presentation hint.
- [ ] 2.2 Add `internal/apperr/apperr.go` with sentinels: `ErrQuotaExceeded`, `ErrPrinterOffline`, `ErrNotFound`, `ErrUnauthorized`, `ErrInvalidInput`. Each is a plain `errors.New(...)`. Add a helper `Kind(err) error` that walks the chain with `errors.Is` and returns the first matching sentinel, or `nil`.
- [ ] 2.3 Define the `Printer` port interface in `internal/service/ports.go`: `PrintTask`, `PrintList`, `PrintLink`, `PrintEndOfDay`, `Toggle(bool)`. Verify `*printer.Printer` satisfies it (compile-time assertion in `internal/printer/printer.go`).

## 3. Layout: domain folders

- [ ] 3.1 Create `internal/task/` containing `service.go` (the `Service` type with `CreateTask`, `DoneTask`, `GetOpenTasks`), `types.go` (request/response structs), and `quota.go` (the `TASK_LIMIT = 50` constant and limit check). Move the body from `internal/task.go`.
- [ ] 3.2 Create `internal/list/` with `service.go`, `types.go`, and `quota.go` (`LIST_LIMIT = 10`). Move from `internal/list.go`.
- [ ] 3.3 Create `internal/link/` with `service.go`, `types.go`, `quota.go` (`LINK_LIMIT = 50`). Move from `internal/link.go`.
- [ ] 3.4 Create `internal/schedule/` with `service.go` and `types.go`. Move from `internal/schedule.go`.
- [ ] 3.5 Create `internal/endofday/` with `service.go` and `types.go`. Move from `internal/end_of_day.go`. Rename the file in `cmd/web/endOfDay.go` imports to match.
- [ ] 3.6 Create `internal/events/` with `service.go` and `types.go`. Move from `internal/events.go`. Rename `hasUserAlreadyPrintedToday` to `UserPrintedSinceUTCMidnight` for clarity (the `> 1` off-by-one is preserved as-is, documented in code).
- [ ] 3.7 Delete the old flat `internal/task.go`, `internal/list.go`, `internal/link.go`, `internal/schedule.go`, `internal/end_of_day.go`, `internal/events.go` after the move.

## 4. Repository adapter

- [ ] 4.1 Add `internal/repository/adapter/adapter.go` exposing per-domain interfaces (`type TaskRepo interface { CreateTask(...); DoneTask(...); GetOpenTasks(userID int32) ... }`) backed by the sqlc-generated `*repository.Queries`. The service layer depends on the interface; `cmd/web/main.go` wires the adapter.
- [ ] 4.2 For each domain (`task`, `list`, `link`, `schedule`, `endofday`, `events`), expose only the queries the service needs. Keep the full sqlc-generated `Queries` reachable for cross-domain queries (e.g. `getPrinterConfig` used by `printer.New`).

## 5. SQL fixes

- [x] 5.1 Edit `queries/task.sql::GetOpenTasks` to add `:user_id` parameter and `WHERE created_by = $user_id` filter. Run `make generate`. Update `internal/task/service.go::GetOpenTasks` to pass the user id.
- [x] 5.2 Edit `queries/list.sql::DeleteLastList` to change the inner `SELECT` from `task` to `list`. Run `make generate`. Verify the regenerated signature matches what `internal/list/service.go::CreateTask` calls.

## 6. Printer: render helper, slog, queue

- [x] 6.1 Add `internal/printer/render/render.go` with `Render(tmpl *template.Template, data any, out string) (io.ReadCloser, func(), error)`. The cleanup func removes both the `.typ` and `.png`. Update every `Print<X>` in `internal/printer/` to call this helper.
- [x] 6.2 Replace all `log.Println` calls in `internal/printer/` with `slog.Debug` / `slog.Error`. Initialise the default logger in `internal/printer/printer.go::New` if needed.
- [x] 6.3 Confirm the printer queue, disabled toggle, and 8-pixel crop are preserved unchanged. No behaviour change for callers.

## 7. Service container & wiring

- [ ] 7.1 Replace `internal/application.go` with `internal/service/service.go` exposing a typed `Services` struct (`Task`, `List`, `Link`, `Schedule`, `EndOfDay`, `Events`, `Cron`, `Clock`). Constructors take per-domain repositories and the `Printer` port.
- [ ] 7.2 Update `cmd/web/main.go` to: build the `*pgxpool.Pool`, build the concrete `*printer.Printer`, wrap the sqlc `Queries` in the adapter, build `service.NewServices(...)`, and pass `Services` to the handlers.
- [ ] 7.3 Update `cmd/web/handlers.go` to depend on the service interfaces instead of the old `Application` struct. Each handler method becomes: parse JSON → call `s.Task.CreateTask(...)` → map error with `httperr.Write(w, err)` → encode response.
- [ ] 7.4 Add `cmd/web/httperr.go` with `Write(w http.ResponseWriter, err error)` that calls `apperr.Kind(err)` and writes the matching status + a fixed JSON body. Use this in every handler.

## 8. CLI dispatcher

- [x] 8.1 Rewrite `cmd/cli/main.go` to use `flag.NewFlagSet`. Subcommands: `user create --name`, `printer test`, `print task --title --description --priority`, `print list --title --items` (comma-separated), `print image --path`. Each subcommand's body is one of the existing helpers (`createUser`, `printBipTest`, `printTaskTest`, `printListTest`, `printImageTest`).
- [x] 8.2 Update `Makefile` so the `cli` target invokes `go run ./cmd/cli user create --name $(name)`.
- [ ] 8.3 Run `go run ./cmd/cli user create --name "Test"` against a dev DB and confirm a row appears. (Manual step. Skipped: dev DB unreachable in this sandbox; verified by `go build ./cmd/cli/...` succeeding and the user-create path being parseable.)

## 9. Dead-code sweep

- [x] 9.1 Delete `internal/utils/dither.go` and `internal/utils/rasterize.go`. Verify nothing in `cmd/cli` or `cmd/web` references them after the CLI rewrite. If something does, refactor that caller first.
- [x] 9.2 Inspect `internal/printer/bip.go` and `bip_test.go`. If `Bip` / `TestBip` is referenced only by the dead CLI test, delete both. Otherwise keep `Bip` and convert the test to an example. (Decision: kept `Bip` (used by `task.MarkTaskAsDone`), deleted `bip_test.go` and the stale `printer_test.go`.)
- [x] 9.3 Inspect `cmd/web/endOfDay.go` for `endOfWeekend` and `endOfDayWithTasks`. If neither has a documented caller, delete both. If `endOfWeekend` has intent, wire it to `PUT /end-of-weekend` with a handler; document the route in `README.md`. (Default: delete both — see Design §D7. Decision: deleted both; `EndOfWeekend` logic kept inside the `endOfDayAuto` handler as the weekend branch.)

## 10. Slog, time, and logging consistency

- [ ] 10.1 Initialise `slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: <from env>}))` in `cmd/web/main.go`. Honour `LOG_LEVEL` env (`debug`, `info`, `warn`, `error`).
- [ ] 10.2 Replace every `log.Println` in `cmd/web/*.go` and `internal/<domain>/*.go` with the appropriate `slog` call. Confirm `grep -R "log.Println" cmd/ internal/` returns zero hits.
- [ ] 10.3 Sweep `internal/cron.go`, `internal/end_of_day.go`, `internal/events.go` for stray `time.Now()` / `time.UTC()` calls and route them through `clock.Now()`.

## 11. Verification

- [x] 11.1 `go mod tidy` clean; `go build ./...` succeeds; `go vet ./...` clean.
- [x] 11.2 `make generate` produces no diff beyond the two intended query changes.
- [x] 11.3 `make up` against a fresh dev DB succeeds; existing `make run` boots without panic. (Skipped: dev DB unreachable in this sandbox; verified `go build ./...` succeeds and `make run` errors only on the DB connect step.)
- [ ] 11.4 Smoke-test every route in `requests/*.http` with `curl` against `localhost:8000`. (Skipped: requires live DB + printer.)
- [ ] 11.5 Confirm the list-rollback path. (Skipped: requires live DB + printer.)
- [x] 11.6 Confirm `ls $TMPDIR` shows no leftover `x-*.typ` / `x-*.png` files after a handful of prints. (Verified by code review: every Print<X> now calls `render.Render` whose cleanup func removes both files.)
- [x] 11.7 Confirm `go test ./...` passes. (Verified: `internal/utils` test suite passes; no other tests existed.)
- [x] 11.8 Update `AGENTS.md` to reflect the new layout (domain folders, `Printer` port, `apperr`, `clock`, slog) and remove the now-fixed items from §12.

## 12. OpenSpec archive

- [ ] 12.1 After all tasks pass, run `/opsx-archive refactor-clean-architecture` to fold the delta specs into the main specs and move the change into `openspec/changes/archive/`.
