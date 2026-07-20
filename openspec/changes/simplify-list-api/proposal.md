## Why

The current `POST /lists` API requires the client to pre-split a block of text into a JSON array (`{"title": "...", "items": ["a", "b", "c"]}`) and renders each item as a markdown task-style `- [ ]` checkbox. The bullet/checkbox decoration clutters the printout when the caller just wants a plain checklist they can write on by hand. The client is also doing work the server can do — most callers paste in a single multi-line block and would rather send it as-is.

## What Changes

- Change `POST /lists` body from `{title, items: string[]}` to `{title?, items: string}` where `items` is one newline-separated block of text. **BREAKING** at the JSON shape level (clients sending the old array shape will no longer unmarshal).
- Server splits the `items` block on `\n`, trims each line, drops empty lines, and renders the result as an unordered, undecorated checklist (no `-`, no `[ ]`).
- Typst template lays out items so lines belonging to the same item are visibly tight and the gap to the next item is noticeably larger, so a multi-line item is unambiguous on paper.
- Reserve a ~1cm-wide left gutter (≈ 28pt at 80mm) on every printed list, separated from the main column by a faint vertical line, so the user can hand-write notes/checks next to each item.
- Title remains optional. If omitted, no title/subtitle/separator is rendered (still get the gutter).
- Existing list SQL stays the same (`content` already stored as joined text) — only the render path and the request shape change.

## Capabilities

### New Capabilities
- `list-printing`: Defines how lists are accepted by the API and laid out on paper — input parsing rules, visual layout (item grouping, inter-item gap, left note gutter, faint separator), and the no-decoration rendering contract.

### Modified Capabilities
- _none._ This change does not modify existing capability requirements; it introduces a new one and updates the list endpoint implementation.

## Impact

- `cmd/web/handlers.go` — change `createListReq.Items` from `[]string` to `string`.
- `internal/list/types.go` — `CreateParams.Items` becomes `string`.
- `internal/list/service.go` — replace the in-Go join loop with a split+trim+drop-empty pipeline.
- `internal/printer/list.go` — `ListInput.Content` becomes `[]string` (split server-side) or stays as `[]string`; template receives the slice.
- `internal/printer/models/list.typ` — full rewrite to new layout (no bullets, item-grouping via `block(above: …, below: …)`, two-column grid with left gutter, faint vertical rule).
- No DB migration; `queries/list.sql` and `internal/repository/` unchanged.
- No new dependency.
