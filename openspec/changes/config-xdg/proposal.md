## Why

The CLI currently requires callers to provide the printer IP through a flag or environment variable, while its other print commands read a separate database-backed configuration. A local, standard configuration file gives every CLI command one persistent printer target without requiring a running database or shell setup.

## What Changes

- Add a TOML configuration file for the CLI at the XDG configuration location: `$XDG_CONFIG_HOME/ticket-control/config.toml`, or `~/.config/ticket-control/config.toml` when the environment variable is unset.
- Load the configuration once at CLI startup into an application struct shared by CLI commands.
- Configure the printer IP through `[printer].ip` and expose it to every CLI command through the application struct.
- Preserve explicit `--ip` and `PRINTER_IP` overrides for one-off and scripted runs.
- Use the configured IP as the default direct target for the existing `typewriter` command.

## Capabilities

### New Capabilities
- `cli-configuration`: Resolve, validate, and expose a TOML-based XDG configuration to all CLI commands.
- `cli-printer-connection`: Connect the existing direct typewriter client to the printer IP supplied by CLI configuration or explicit overrides.

### Modified Capabilities

- None.

## Impact

- Affects `cmd/cli`, especially command dispatch and typewriter configuration.
- Adds a Go TOML parsing dependency because the standard library does not parse TOML.
- Does not change web-server printer configuration, database schema, API contracts, or database-backed CLI print commands.
