-- name: CreateScheduleTask :exec
INSERT INTO schedule (name, title, description, cron_expression, created_by)
VALUES ($1, $2, $3, $4, $5);

-- name: ToggleScheduleTask :one
UPDATE schedule
SET enabled = NOT enabled, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND created_by = $2
RETURNING id, name, title, description, cron_expression, enabled, created_at, updated_at;

-- name: GetUserScheduleTasks :many
SELECT id, name, title, description, cron_expression, enabled, created_at, updated_at
FROM schedule
WHERE created_by = $1
ORDER BY created_at DESC;

-- name: GetAllEnabledScheduleTasks :many
SELECT id, name, title, description, cron_expression, enabled, created_by
FROM schedule
WHERE enabled = TRUE
ORDER BY created_at DESC;
