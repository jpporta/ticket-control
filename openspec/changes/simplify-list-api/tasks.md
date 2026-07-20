## 1. Wire-format change (HTTP + service boundary)

- [x] 1.1 In `cmd/web/handlers.go`, change `createListReq.Items` from `[]string` to `string` and add a small helper that splits on `\n`, trims each line, and drops empties; return `apperr.ErrInvalidInput` if the result is empty.
- [x] 1.2 Pass the resulting `[]string` into `list.CreateParams{Items: ...}` exactly as today.
- [x] 1.3 In `internal/list/service.go::CreateList`, remove the in-Go `+= item + "\n"` join; keep `Items []string` as the domain shape and join into the storage string at the repo boundary via `strings.Join(items, "\n")`.
- [x] 1.4 Update `requests/createList.http` to the new body shape (`{"title": "...", "items": "a\nb\nc"}`) and add a second example without a title.

## 2. Typst layout rewrite

- [x] 2.1 Rewrite `internal/printer/models/list.typ`: drop `- [ ]`, set `set par(leading: 0.55em)`, wrap each item in `#block(above: 0pt, below: 1.1em, ...)`, and pad the content with `#pad(left: 28pt, ...)`.
- [x] 2.2 Add `#place(top + left, dx: 28pt, dy: 0pt, line(length: 100%, stroke: gray + 0.4pt))` to draw the faint vertical separator at 28pt from the left edge.
- [x] 2.3 Make the title/subtitle/`#line(length: 50%)` block conditional on `.Title` being non-empty (omit entirely when blank).
- [x] 2.4 Smoke-test the template with a sample `PrintList` payload via `make run` + curl; if a `make typst-list` target doesn't exist, add one mirroring `make typst-task` for iteration.

## 3. Verify no other callers broke

- [x] 3.1 Grep `internal/` and `cmd/` for any other call site of `list.Service.CreateList` or `CreateListReq`; update them to the new shape (expected: none — `handlers.go` is the only caller).
- [x] 3.2 Confirm `internal/printer/list.go::PrintList` still receives a `[]string` and feeds it into the template unchanged.
- [x] 3.3 Confirm `queries/list.sql` is unchanged and no `make generate` run is needed.

## 4. Manual smoke test

- [ ] 4.1 `make up` (no-op expected — no migration in this change) and `make run`.
- [ ] 4.2 `curl -X POST /lists` with the new shape, observe the printed ticket has no bullets, items are spaced apart, and a faint vertical line sits ~1cm from the left edge.
- [ ] 4.3 Send a titleless list — verify the title block is absent but the gutter and separator remain.
- [ ] 4.4 Send `{"items": "\n\n   \n"}` — verify HTTP 400, no DB row, no paper.
- [ ] 4.5 Send `{"items": ["a","b"]}` (old shape) — verify HTTP 400.
- [ ] 4.6 Toggle the printer off, send a list, verify the request errors with `apperr.ErrPrinterOffline` (queue-on-disabled still works), and the DB row is rolled back.
