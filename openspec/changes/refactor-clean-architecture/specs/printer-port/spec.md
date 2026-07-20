## ADDED Requirements

### Requirement: `Printer` is an interface owned by the service layer
The `Printer` port MUST be defined as a Go interface in the service-facing package (not inside `internal/printer`). The interface MUST include `PrintTask`, `PrintList`, `PrintLink`, `PrintEndOfDay`, and `Toggle(bool)`. The concrete `*printer.Printer` in `internal/printer` MUST satisfy the interface; `cmd/web/main.go` is the only wiring site.

#### Scenario: Service code references only the interface
- **WHEN** a service method calls `p.PrintTask(...)`
- **THEN** `p` is typed as the interface; no concrete `*printer.Printer` is imported in the service package

#### Scenario: Concrete printer can be substituted in tests
- **WHEN** a service test provides a struct implementing only the methods it exercises
- **THEN** the service code compiles and the test runs without TCP, escpos, or Typst

### Requirement: Printer queueing is internal
The "disabled printer queues closures, drains when re-enabled" behaviour MUST be preserved and MUST remain internal to the concrete printer. Callers MUST only see `Print<X>` returning an error when the printer is disabled and `Toggle(true)` flushing the queue.

#### Scenario: Disabled printer accepts no immediate print
- **WHEN** `Toggle(false)` has been called and a service calls `PrintTask`
- **THEN** the call returns `ErrPrinterOffline` and the closure is enqueued internally

#### Scenario: Re-enabled printer drains the queue
- **WHEN** `Toggle(true)` is called after a period of being disabled with queued closures
- **THEN** each queued closure is executed in order, with a small delay between jobs; if any closure fails, the error is logged and the drain continues

### Requirement: 8-pixel height crop is preserved
Every `Print<X>` method MUST crop the rendered PNG height to a multiple of 8 pixels before sending it to the printer. This requirement exists because the ESC/POS raster command rejects non-aligned heights.

#### Scenario: Any printable image is height-aligned
- **WHEN** a Typst template renders a PNG whose height is not a multiple of 8
- **THEN** the image sent to the printer has its bottom padded to the next multiple of 8; no ESC/POS error is raised
