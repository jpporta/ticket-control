## ADDED Requirements

### Requirement: List API accepts a newline-separated items block
The `POST /lists` endpoint MUST accept a request body with an optional `title` (string) and a required `items` field whose value is a single string. The server MUST split `items` on `\n`, trim each resulting line, and drop lines that are empty after trimming. The endpoint MUST reject requests where the trimmed result has zero items with HTTP 400 (`apperr.ErrInvalidInput`).

#### Scenario: Single-line block creates one item
- **WHEN** a client sends `{"items": "buy milk"}`
- **THEN** the server persists and prints a list containing exactly one item: `buy milk`

#### Scenario: Multi-line block creates one item per non-empty line
- **WHEN** a client sends `{"items": "buy milk\npay rent\ncall dentist"}`
- **THEN** the server persists and prints three items, in order: `buy milk`, `pay rent`, `call dentist`

#### Scenario: Empty and whitespace-only lines are dropped
- **WHEN** a client sends `{"items": "first\n\n   \nsecond\n"}`
- **THEN** the server persists and prints two items: `first` and `second` (the empty line and the whitespace-only line are discarded)

#### Scenario: Title is optional
- **WHEN** a client sends `{"items": "a\nb\nc"}` (no `title`)
- **THEN** the server accepts the request and prints the list without a title heading

#### Scenario: Empty items after trimming is rejected
- **WHEN** a client sends `{"items": "\n\n   \n"}` (only empty/whitespace lines)
- **THEN** the server returns HTTP 400 and does not persist or print anything

#### Scenario: Legacy array shape is rejected
- **WHEN** a client sends `{"items": ["a", "b"]}` (the old array shape)
- **THEN** the server returns HTTP 400 because the JSON unmarshals `items` as an array, not a string

### Requirement: Printed list has no list-style decoration
The printed list MUST NOT render any bullet character, checkbox (`[ ]`/`[x]`), or numeric marker on items. Items MUST be plain text lines.

#### Scenario: No bullet or checkbox appears on any item
- **WHEN** the server renders a list with items `["a", "b"]`
- **THEN** the rendered output contains neither `- ` nor `[ ]` nor any other decoration prefix on the items

### Requirement: Items are visually grouped; inter-item gap exceeds intra-item line spacing
The Typst layout MUST make lines that belong to the same item sit visibly closer together than the gap between two distinct items. A multi-line item MUST be unambiguous on paper: every line of a given item is closer to its sibling lines than to the first line of the next item.

#### Scenario: Single-line items still get the inter-item gap
- **WHEN** the server renders `["a", "b", "c"]`
- **THEN** the vertical distance between `a` and `b` equals the distance between `b` and `c`, and is greater than the baseline line-height of a single item

#### Scenario: Multi-line item is grouped
- **WHEN** the server renders `["first line\nsecond line", "next item"]`
- **THEN** the line containing `second line` is visually closer to the line containing `first line` than to the line containing `next item`

### Requirement: Reserved left gutter with faint vertical separator
Every printed list MUST reserve a left-side note gutter of approximately 28pt (≈1cm at the printer's dot pitch) of horizontal space, intended for handwritten notes or checks. A faint vertical line (color `gray`, thickness ≈0.4pt) MUST separate the gutter from the main content column. The gutter and separator MUST appear on every list, regardless of whether a title is present.

#### Scenario: Gutter and separator appear on a titled list
- **WHEN** the server renders a list with a title
- **THEN** the printed output shows the title and items aligned to the right of a faint vertical line, with empty space to the left of that line

#### Scenario: Gutter and separator appear on a titleless list
- **WHEN** the server renders a list without a title
- **THEN** the printed output shows the items aligned to the right of a faint vertical line, with empty space to the left of that line

#### Scenario: Separator is faint, not solid black
- **WHEN** the server renders any list
- **THEN** the vertical separator is rendered in `gray` at ≈0.4pt (not solid black and not absent)

### Requirement: Quota, persistence, and print-failure behavior are unchanged
The list MUST continue to count against the user's daily `ListLimit` (10 lists per UTC day) and MUST continue to follow the print-then-store-with-rollback contract: on print failure the row is deleted and the user is not charged a quota slot for a successful insert that was then rolled back by the quota check. The DB schema, queries, and `list` table columns MUST NOT change.

#### Scenario: Successful print consumes one quota slot
- **WHEN** a user who has printed 9 lists today successfully prints a 10th
- **THEN** the response is `201 Created`, the row is persisted, and the next request for that user in the same UTC day returns HTTP 429 (`apperr.ErrQuotaExceeded`)

#### Scenario: Print failure rolls back the insert and quota slot
- **WHEN** the printer is offline or the print errors out
- **THEN** the `list` row inserted for the request is deleted, no quota slot is consumed, and the caller receives an error mapped through `apperr.ErrPrinterOffline`
