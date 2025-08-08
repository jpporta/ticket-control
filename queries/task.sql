-- name: CreateTask :one
INSERT INTO task (title, description, priority, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetNoUsersTask :one
SELECT count(*) AS total FROM task
WHERE created_by = $1
AND created_at >= $2
AND created_at < $3;

