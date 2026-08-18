# Tasks

- [x] 1. Add dial retry to `internal/printer/printer.go` `start()` (3 attempts, 2s apart)
- [x] 2. Create `cmd/cli/typewriter.go`: flags, env fallback, printer address resolution
- [x] 3. Terminal raw mode via `stty raw -echo` with restore on defer and on SIGTERM/SIGHUP
- [x] 4. ESC/POS byte helpers: reset, select codepage, feed, cut
- [x] 5. CP850/CP437/CP860/Windows-1252 transcoding with drop-on-unmappable
- [x] 6. Line buffer, backspace, echo, dial-per-line commit
- [x] 7. ~~Raw mode~~ — built, tested on hardware, output identical to buffered; removed
- [x] 8. Key handling: Enter, Backspace, Esc (with escape-sequence disambiguation), Ctrl-D, Ctrl-C
- [x] 9. Register `typewriter` in the CLI dispatcher and usage text
- [x] 10. Unit test for transcoding and escape-sequence disambiguation
- [x] 11. `go build ./...` and `go test ./...`
- [x] 12. Print piped stdin as one ticket when stdin is not a terminal
