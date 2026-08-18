# typewriter

## ADDED Requirements

### Requirement: Direct printer session
The CLI SHALL provide a `typewriter` subcommand that connects directly to the
thermal printer over TCP without involving the HTTP server or the database.

#### Scenario: Printer address from flags
- **WHEN** `typewriter` is invoked with `--ip` and `--port`
- **THEN** the client connects to that address

#### Scenario: Printer address from environment
- **WHEN** `--ip` or `--port` are omitted
- **THEN** the client falls back to `PRINTER_IP` and `PRINTER_PORT`

#### Scenario: Unreachable printer
- **WHEN** the printer cannot be reached at session start
- **THEN** the client reports the error and exits without altering the terminal

### Requirement: Line composition
The client SHALL hold the current line locally and keep it editable until
committed.

#### Scenario: Committing a line
- **WHEN** the user presses Enter
- **THEN** the held line is sent to the printer terminated by `LF` and the
  buffer is cleared

#### Scenario: Editing before commit
- **WHEN** the user presses Backspace with a non-empty buffer
- **THEN** the last character is removed from the buffer and from the screen

#### Scenario: Connection lifetime
- **WHEN** a line is committed
- **THEN** the client opens a connection, writes the line, and closes it, so the
  printer's single TCP slot is not held between lines

### Requirement: Piped input
When standard input is not a terminal, the client SHALL print its entire
contents as one ticket and exit.

#### Scenario: Command output piped in
- **WHEN** the subcommand is invoked with stdin connected to a pipe or file
- **THEN** the contents are transcoded, printed, cut, and the client exits
  without entering interactive mode

#### Scenario: Input without a trailing newline
- **WHEN** the piped input does not end in a newline
- **THEN** a newline is appended so the final line is committed to paper

### Requirement: Cutting
The client SHALL feed blank lines before every cut so that text is not severed.

#### Scenario: Cut and continue
- **WHEN** the user presses Esc
- **THEN** the printer feeds and cuts, and the session continues accepting input

#### Scenario: Cut and exit
- **WHEN** the user presses Ctrl-D
- **THEN** the printer feeds and cuts, the connection closes, and the client exits

#### Scenario: Abort without cutting
- **WHEN** the user presses Ctrl-C
- **THEN** the client exits without cutting the paper

#### Scenario: Escape sequence is not a cut
- **WHEN** an `ESC` byte is followed within the disambiguation window by further
  bytes (an arrow key or other escape sequence)
- **THEN** the sequence is discarded and no cut occurs

### Requirement: Text encoding
The client SHALL transcode terminal UTF-8 input to the printer's selected codepage.

#### Scenario: Default codepage
- **WHEN** no `--codepage` is given
- **THEN** CP850 is selected on the printer and used for transcoding

#### Scenario: Unmappable character
- **WHEN** a typed rune has no representation in the selected codepage
- **THEN** the rune is dropped rather than emitted as an incorrect byte

### Requirement: Terminal restoration
The client SHALL restore the terminal to its prior state on every exit path.

#### Scenario: Signal during session
- **WHEN** the process receives `SIGTERM` or `SIGHUP` mid-session
- **THEN** the terminal mode is restored before the process exits

## MODIFIED Requirements

### Requirement: Printer connection
Opening a printer connection SHALL tolerate transient dial failures by retrying
before reporting an error.

#### Scenario: Printer momentarily busy
- **WHEN** a print is attempted while the printer's single TCP slot is occupied
- **THEN** the dial is retried before the operation fails, so short collisions do
  not cause the caller to roll back its work
