## Context

`cmd/cli` dispatches commands as package-level functions. The direct `typewriter` command currently resolves its printer IP from `--ip` or `PRINTER_IP`; other CLI print commands construct the service printer, which reads its configuration from Postgres. The CLI needs one configuration value available from startup without changing the web service or its database-backed printer path.

## Goals / Non-Goals

**Goals:**
- Read one TOML configuration file at CLI startup into an application struct.
- Resolve the file using `os.UserConfigDir`, which honors `XDG_CONFIG_HOME` on Linux and falls back to `~/.config`.
- Make the file's printer IP the normal `typewriter` default while retaining explicit command-line and environment overrides.
- Fail startup with an actionable error when the file is missing, malformed, or lacks `printer.ip`.

**Non-Goals:**
- Do not change web-server configuration or its Postgres printer record.
- Do not migrate `printer test`, `print task`, or `print list` away from the service printer.
- Do not add a CLI command that writes configuration.
- Do not configure printer ports, authentication, or additional CLI settings yet.

## Decisions

### XDG file path

Use `os.UserConfigDir` and append `ticket-control/config.toml`. This follows the platform's configuration convention without manually duplicating XDG fallback rules. A project-specific `~/.xtgconfig` directory was rejected because it is nonstandard and offers no benefit.

### Minimal configuration schema

Decode only:

```toml
[printer]
ip = "192.168.1.50"
```

The existing default port remains `9100`. Keeping the schema to the required IP avoids inventing settings before they are needed.

### Application ownership and precedence

`main` loads configuration once and passes an application struct to command handlers. `typewriter` resolves its target in this order: `--ip`, `PRINTER_IP`, then the loaded TOML IP. Flags remain the highest-priority one-off override; environment variables retain existing script compatibility. Other commands receive the application struct but do not consume the printer IP yet.

### TOML decoder

Add a maintained TOML decoder dependency. Go's standard library has no TOML parser; a restricted handwritten parser would not reliably implement the advertised file format.

## Risks / Trade-offs

- [Missing config prevents even non-printer CLI commands from starting] -> This is intentional for a global startup configuration requirement; the error names the required path and field.
- [Existing scripts depend on `PRINTER_IP`] -> Keep environment variables as a higher-precedence fallback than the file.
- [Future commands need port or other settings] -> Extend the TOML schema only when a command actually needs them; the current port remains the established `9100` default.
