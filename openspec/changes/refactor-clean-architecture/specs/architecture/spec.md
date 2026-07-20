## ADDED Requirements

### Requirement: Layered package layout
The codebase MUST organise code into four layers with strictly one-way dependencies: `cmd` → `service` → `repository` → `printer`. Each layer depends only on the layer immediately below it. Domain logic MUST live in `internal/<domain>/` packages (one folder per bounded context), not in flat `internal/*.go` files.

#### Scenario: New domain lands in the right place
- **WHEN** a developer adds a new bounded context (for example, "notes")
- **THEN** they create `internal/notes/` containing the service type and request/response structs, with no direct import of `cmd/web` or `internal/printer`'s concrete types

#### Scenario: Handler imports only the service layer
- **WHEN** an HTTP handler in `cmd/web/` is wired
- **THEN** it imports the relevant `internal/<domain>` service interface and the central error mapper; it does not import `internal/repository` or `internal/printer`

### Requirement: Service depends on a Printer port
The `Printer` port MUST be a Go interface defined in the service-facing package (not in `internal/printer`). Services MUST depend on the interface; `cmd/web/main.go` is the only place that wires the concrete `*printer.Printer` to it. The interface MUST expose at least: `PrintTask`, `PrintList`, `PrintLink`, `PrintEndOfDay`, `Toggle(bool)`. The queue, disabled-state, and TCP/Typst machinery stay internal to the concrete type.

#### Scenario: Service test substitutes a stub printer
- **WHEN** a service test calls `CreateTask` without a real printer
- **THEN** it can supply a struct that satisfies the `Printer` interface with no-op or recording methods, and the test runs without network access

#### Scenario: Adding a new printable type
- **WHEN** a developer adds `PrintReceipt` to the concrete printer
- **THEN** the interface gains one method; existing callers do not break because they only call methods they already use

### Requirement: Central error model
The system MUST expose sentinel errors in `internal/apperr` for at least: `ErrQuotaExceeded`, `ErrPrinterOffline`, `ErrNotFound`, `ErrUnauthorized`, `ErrInvalidInput`. Services MUST return errors that wrap these sentinels via `fmt.Errorf("...: %w", apperr.ErrX)`. A single HTTP helper in `cmd/web` MUST walk the error chain with `errors.Is` and map each sentinel to a fixed HTTP status code (429 / 503 / 404 / 401 / 400 respectively) without leaking internal error text in the response body.

#### Scenario: Quota exceeded returns 429
- **WHEN** a user has hit the daily task limit and posts a new task
- **THEN** the HTTP response is 429 with a fixed JSON body and the underlying `ErrQuotaExceeded` sentinel is reachable via `errors.Is`

#### Scenario: Printer offline returns 503
- **WHEN** the printer is disabled and a print is requested
- **THEN** the HTTP response is 503 with a fixed JSON body and the response body does not contain the raw wrapped error message

#### Scenario: Unknown error returns 500 without leaking internals
- **WHEN** an unanticipated error reaches the handler
- **THEN** the response is 500 with a generic body; the underlying error is logged with `slog.Error` including the request path and user id

### Requirement: Single clock source
The system MUST read wall-clock time through `internal/clock.Now()` returning `time.Now().UTC()`. Quota windows, "already printed today" checks, and cron predicates MUST call this function. Direct `time.Now()` calls in domain code are forbidden outside `internal/clock` itself.

#### Scenario: Quota window is UTC regardless of server TZ
- **WHEN** the server's local timezone is not UTC and a user crosses 00:00 local time
- **THEN** the quota window rolls over at 00:00 UTC, not at local midnight

### Requirement: Temp-file cleanup is mandatory
Every `Print<X>` method MUST go through a `render` helper that owns the `.typ` and `.png` temp-file lifecycle. The helper MUST remove both files before returning. Hand-rolled `os.CreateTemp` calls in `Print<X>` methods are forbidden.

#### Scenario: Print leaves no temp files
- **WHEN** a `PrintTask` call completes (success or failure)
- **THEN** `$TMPDIR` contains no leftover `x-*.typ` or `x-*.png` files from that call

### Requirement: Structured logging
The service MUST use `log/slog` with a JSON handler. `log.Println` calls in production code MUST be replaced with `slog.Info` / `slog.Error` carrying structured attributes (`slog.String("path", r.URL.Path)`, etc.). Log level MUST be configurable via the `LOG_LEVEL` environment variable.

#### Scenario: Request is logged with structured fields
- **WHEN** any authenticated request is processed
- **THEN** `slog.Info` is called with `path`, `method`, `user_id`, and `status` as structured attributes; no `log.Println` calls remain in the request path

### Requirement: CLI is a real subcommand dispatcher
`cmd/cli/main.go` MUST parse subcommands via `flag.NewFlagSet` and dispatch to: `user create --name`, `printer test`, `print task --title --description --priority`, `print list --title --items`, `print image --path`. The existing helper functions become the bodies of these subcommands. The `make cli name="X"` Make target MUST translate to `go run ./cmd/cli user create --name "X"`.

#### Scenario: `make cli name="X"` creates a user
- **WHEN** a developer runs `make cli name="Test User"`
- **THEN** a row is inserted into the `user` table with a generated API key, and the key is printed to stdout

#### Scenario: `go run ./cmd/cli printer test` prints a test page
- **WHEN** a developer runs the printer test subcommand
- **THEN** the existing `printBipTest` helper is invoked and a test page is sent to the printer

### Requirement: Quiet hours and quota numbers preserved
The internal cron job MUST keep the existing schedule: disable the printer at 22:00 and enable it at 08:00 (server-local). Per-user daily quotas MUST keep their existing values: tasks 50, lists 10, links 50. These numbers MUST be defined as named constants in one location per domain, not scattered.

#### Scenario: Quiet hours match existing schedule
- **WHEN** the server clock reaches 22:00
- **THEN** the printer is disabled via `printer.Toggle(false)` exactly as it is today

#### Scenario: Quota constants are referenced in one place
- **WHEN** a developer greps for `TASK_LIMIT` or `LIST_LIMIT`
- **THEN** there is exactly one definition per constant in `internal/task/` and `internal/list/` respectively
