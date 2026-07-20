## ADDED Requirements

### Requirement: `GetOpenTasks` filters by the requesting user
The `GetOpenTasks` query MUST accept a `created_by` parameter and MUST return only rows whose `created_by` matches the parameter. The service-layer method MUST pass the authenticated user's id into the query. The handler MUST continue to expose `GET /task` and require the `X-Api-Key` header.

#### Scenario: User A cannot see User B's open tasks
- **WHEN** User A calls `GET /task`
- **THEN** the response contains only tasks where `created_by = A`; tasks created by User B are excluded

#### Scenario: User with no tasks gets an empty list
- **WHEN** a user with no open tasks calls `GET /task`
- **THEN** the response is `200 OK` with an empty JSON array

### Requirement: `DeleteLastList` targets the `list` table
The `DeleteLastList` query MUST delete a row from the `list` table, not from `task`. The inner `SELECT` MUST reference `list`. The service-layer rollback path MUST call this query with the id of the list row that was just inserted, and the row MUST be removed on rollback.

#### Scenario: List rollback deletes the inserted list row
- **WHEN** a list is inserted and the subsequent print fails
- **THEN** the rollback deletes the list row from the `list` table by the inserted id; no `task` row is affected

#### Scenario: Successful list print leaves both tables intact
- **WHEN** a list is inserted and the print succeeds
- **THEN** the list row remains in the `list` table and no rollback is issued
