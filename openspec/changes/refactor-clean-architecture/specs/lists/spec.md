## MODIFIED Requirements

### Requirement: `DeleteLastList` targets the `list` table
The `DeleteLastList` query MUST delete a row from the `list` table, not from `task`. The inner `SELECT` MUST reference `list`. The service-layer rollback path MUST call this query with the id of the list row that was just inserted, and the row MUST be removed on rollback.

> **Note:** This is the same requirement as `specs/tasks/spec.md` for the shared `lists` capability. It is repeated here as a delta against the existing `lists` capability so the archive step can preserve the historical delta.

#### Scenario: List rollback deletes the inserted list row
- **WHEN** a list is inserted and the subsequent print fails
- **THEN** the rollback deletes the list row from the `list` table by the inserted id; no `task` row is affected
