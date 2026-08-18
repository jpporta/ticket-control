## 1. CLI Configuration

- [x] 1.1 Add the TOML decoder dependency and define the CLI application/configuration types.
- [x] 1.2 Resolve and load `$XDG_CONFIG_HOME/ticket-control/config.toml` with validation and actionable startup errors.
- [x] 1.3 Construct the application once in `main` and pass it to CLI command handlers.

## 2. Printer Target

- [x] 2.1 Update `typewriter` to use the application configuration after `--ip` and `PRINTER_IP`, retaining port `9100` as the default.

## 3. Verification

- [x] 3.1 Add focused tests for XDG path resolution, TOML validation, and typewriter IP precedence.
- [x] 3.2 Run formatting and `go test ./...`.
