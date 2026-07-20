## Context

The list endpoint currently accepts `{"title": "...", "items": ["a", "b", "c"]}` and renders each entry as `- [ ] item` in `internal/printer/models/list.typ`. The `- [ ]` decoration makes the printout look like a markdown todo list, which clutters the visual field when the caller just wants a plain list to write on. The client also has to pre-split text into an array, which is friction most callers don't want.

This change keeps the endpoint path (`POST /lists`), the DB schema (`list` table), the quota (`ListLimit = 10`), and the print-then-store-with-rollback flow untouched. It only changes the request shape, the server-side parsing, and the Typst template.

The 80mm print head is ~227pt wide. We reserve a left note gutter of ~28pt (≈1cm at the printer's ~203dpi), then a 0.4pt faint vertical rule, then the main content column.

## Goals / Non-Goals

**Goals:**
- Accept `{title?, items: "<multi-line block>"}` and split server-side on `\n`.
- Render the list with no bullet, no checkbox, no list-style decoration.
- Group lines belonging to the same item tightly; add a clearly larger gap between items.
- Always reserve a ~28pt left gutter for handwritten notes, separated by a faint vertical rule.
- Title stays optional; when absent, no title/subtitle/divider is rendered (gutter remains).
- Stay on the existing ESC/POS + Typst pipeline. No new dependency, no migration, no SQL change.

**Non-Goals:**
- Preserving the old `items: []string` request shape (it's a breaking change, called out in the proposal).
- Adding item-level metadata (priority, done flag, due date) — that's a separate change.
- Allowing CRLF, comma, or other non-`\n` separators. `\n` is the only split token.
- Changing the `list` table, the `content` storage format, or the existing query contract.

## Decisions

### Decision 1 — Split on `\n` only, not `\r\n`
**Choice:** Use `strings.Split(s, "\n")` and `strings.TrimSpace` per line. Drop lines whose trimmed result is empty.
**Why:** Keeps the contract dead simple — the user said "separated by a new line". Real-world clients (Apple Shortcuts, curl, jq) emit `\n`; `\r\n` would only show up if someone passes a Windows-pasted block, and `TrimSpace` already strips the `\r`. No need to normalize further.
**Alternatives:** `bufio.Scanner` — overkill for a one-shot split. `regexp.Split` — unneeded complexity for a literal newline.

### Decision 2 — Server keeps `Items []string` in the domain; only the HTTP layer changes shape
**Choice:** `cmd/web/handlers.go` parses the string into `[]string` (split + trim + drop-empty), then calls `list.Service.CreateList(ctx, list.CreateParams{Title, Items: []string}, ...)` exactly as today. `internal/list/service.go` no longer joins back into a single string — it passes the slice straight to `printer.PrintList`. `queries/list.sql` `CreateList` already takes `content string`, so the service joins there with `"\n".join(items)` for storage.
**Why:** The service-layer contract (`Items []string`) is the right shape internally — printer and DB both want ordered items. Only the wire format and storage serialization change.
**Alternatives:** Push the string all the way down to the printer and let Typst/Go do the split there — leaks parsing concerns into the render layer. Rejected.

### Decision 3 — Typst layout: two-column table with a bordered note gutter
**Choice:**
- Page width stays at `300pt` (matches existing 80mm).
- Set `set par(leading: 0.55em)` for tight in-item line spacing.
- Render each item as one row in a two-column table: an empty `30%` note gutter and a `70%` content column.
- Give each row `1.1em` bottom inset and draw only inner table borders: each non-first row's top edge and the second column's left edge, using `black + 0.6pt` so the stroke survives PNG rasterization and ESC/POS printing.
- Use the installed font family name `BerkeleyMono Nerd Font Mono` and leave `12pt` below the title divider.
- No list marker, no `- [ ]`, no heading decoration on items.
**Why:** A table makes the separator share the exact height of every item row, unlike the previous absolute `place` overlay that disappeared on the auto-height page. It also keeps gutter width, separator, and row spacing in one layout primitive.

### Decision 4 — Faint vertical line via the note-column border
**Choice:** Set the first table column's right border to `gray + 0.4pt`.
**Why:** The border is guaranteed to span every item row and remains faint on thermal paper.

### Decision 5 — ~28pt left gutter, not exactly 1cm
**Choice:** 28pt ≈ 9.9mm on a 72dpi baseline, but Typst uses 72.27pt = 1in, so 28pt = 9.85mm. Close enough to "around 1cm" and cleanly matches the existing template's `width: 300pt` page minus ~272pt content (plenty for the 80mm print head).
**Why:** The exact mm figure doesn't matter for handwritten notes; what matters is that it's wide enough for a checkbox or short word. 28pt gives roughly 7–8mm of usable scribble space, which is enough for "✓", "X", or a 3-letter annotation.
**Alternatives:** 36pt (≈12.7mm, true 1cm at 72dpi) — eats too much of the 300pt page on a long list. 20pt — too narrow for a pen stroke. 28pt is the sweet spot.

### Decision 6 — Empty/whitespace-only `items` returns `apperr.ErrInvalidInput`
**Choice:** If `split + trim + drop-empty` produces zero items, return `apperr.ErrInvalidInput` (400). The handler maps that to HTTP 400 via `cmd/web/httperr.go`.
**Why:** A list with zero items is almost always a client bug; we shouldn't burn a quota slot or a piece of paper on it.
**Alternatives:** Render an empty list — wastes paper and quota. Auto-skip the call — hides the bug.

## Risks / Trade-offs

- **[Breaking wire format]** Existing clients sending `items: ["a","b"]` will get a 400. → Migration plan: update `requests/createList.http` and any Shortcuts/IFFTTT/etc. callers in the same change. Documented in proposal.md.
- **[Typst template breakage]** `internal/printer/text.go` parses the template at startup; a syntax error will fail fast but won't surface until `make run` first prints a list. → Mitigation: keep the template small, eyeball-test with `make typst-list` (analogous to `make typst-task`) before the first print.
- **[Item-vs-line ambiguity if leading whitespace varies]** A user pasting `  foo\n  bar` (intended as one wrapped item) becomes two items after trim. → Acceptable: matches "split on \n + trim" semantics, and the new wide inter-item gap makes the split visually obvious.
- **[Storage format unchanged]** `list.content` is still the `\n`-joined form, so any future direct-DB reads stay backwards-compatible. No data migration.
- **[No new dependency]** Zero supply-chain surface added.

## Migration Plan

1. Land the Go + Typst changes in a single commit.
2. Update `requests/createList.http` to the new shape (`items: "a\nb\nc"`).
3. Manual smoke test: `make run` + `curl -X POST` with the new body.
4. Rollback: revert the commit. No DB migration to undo, no quota impact.

## Open Questions

- None blocking. One nice-to-have: should empty-after-split be `ErrInvalidInput` or silently render an empty page with just the title/gutter? Currently `ErrInvalidInput`; happy to flip if you'd rather.
