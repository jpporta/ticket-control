## ADDED Requirements

### Requirement: CLI loads XDG TOML configuration at startup
The CLI SHALL resolve its configuration file as `ticket-control/config.toml` below the directory returned by `os.UserConfigDir` and SHALL decode it before dispatching a command. The configuration SHALL expose `[printer].ip` through the CLI application struct.

#### Scenario: XDG configuration directory is set
- **WHEN** `XDG_CONFIG_HOME` is set and the CLI starts
- **THEN** the CLI reads `$XDG_CONFIG_HOME/ticket-control/config.toml`

#### Scenario: XDG configuration directory is unset
- **WHEN** `XDG_CONFIG_HOME` is unset and the CLI starts on Linux
- **THEN** the CLI reads `~/.config/ticket-control/config.toml`

#### Scenario: Valid printer configuration
- **WHEN** the configuration contains `[printer]` with a non-empty `ip`
- **THEN** the CLI makes that IP available to every command through its application struct

### Requirement: CLI reports invalid required configuration
The CLI MUST terminate before command dispatch with an actionable error when the configuration file cannot be read, cannot be decoded as TOML, or does not contain a non-empty `printer.ip` value.

#### Scenario: Configuration file is absent
- **WHEN** the resolved configuration file does not exist
- **THEN** the CLI reports the expected configuration path and exits without running a command

#### Scenario: Printer IP is missing
- **WHEN** the configuration lacks a non-empty `printer.ip`
- **THEN** the CLI reports that `printer.ip` is required and exits without running a command
