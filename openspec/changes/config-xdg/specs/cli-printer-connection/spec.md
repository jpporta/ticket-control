## ADDED Requirements

### Requirement: Typewriter uses configured printer IP
The typewriter command SHALL use the printer IP loaded in the CLI application configuration when no explicit IP override is supplied. It SHALL keep port `9100` as the default when no port is supplied.

#### Scenario: Configuration supplies the typewriter target
- **WHEN** `typewriter` runs without `--ip` or `PRINTER_IP`
- **THEN** it connects to the `printer.ip` value loaded from the configuration file

### Requirement: Typewriter supports explicit IP overrides
The typewriter command SHALL resolve its IP in this order: `--ip`, `PRINTER_IP`, then `printer.ip` from configuration.

#### Scenario: Flag overrides configuration
- **WHEN** `typewriter` is invoked with `--ip` and the configuration contains a different printer IP
- **THEN** it uses the `--ip` value

#### Scenario: Environment overrides configuration
- **WHEN** `typewriter` has no `--ip`, `PRINTER_IP` is set, and the configuration contains a different printer IP
- **THEN** it uses the `PRINTER_IP` value
