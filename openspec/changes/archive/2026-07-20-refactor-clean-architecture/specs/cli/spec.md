## ADDED Requirements

### Requirement: `cmd/cli` exposes real subcommands
The CLI binary MUST dispatch on subcommands rather than printing the current time. Subcommands MUST include `user create`, `printer test`, `print task`, `print list`, `print image`. The existing helper functions (`createUser`, `printBipTest`, `printTaskTest`, `printListTest`, `printImageTest`) MUST be called by the subcommand handlers.

#### Scenario: `go run ./cmd/cli user create --name "X"` creates a user
- **WHEN** the user subcommand is invoked with `--name`
- **THEN** the existing `createUser` helper runs, inserts a row into the `user` table, and prints the generated API key

#### Scenario: Unknown subcommand prints usage
- **WHEN** the CLI is invoked with an unrecognised subcommand
- **THEN** usage information is printed to stderr and the process exits non-zero

### Requirement: Makefile `cli` target wires to the user subcommand
The `make cli name="X"` Makefile target MUST invoke `go run ./cmd/cli user create --name "X"`.

#### Scenario: `make cli name="Test User"` creates a user
- **WHEN** the developer runs the target with a quoted name argument
- **THEN** a new user row exists in the database and the API key is visible in the make output
