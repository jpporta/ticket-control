-- name: CreateTask :one
INSERT INTO task (title, description, priority, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetNoUsersTask :one
SELECT count(*) AS total FROM task
WHERE created_by = $1
AND created_at >= $2
AND created_at < $3;

-- name: DeleteLastTask :exec
DELETE FROM task
WHERE id = (
	SELECT id FROM task as t
	WHERE t.created_by = $1
	ORDER BY created_at DESC
	LIMIT 1
);

-- name: GetOpenTasks :many
SELECT id, title, priority, created_at
FROM task
WHERE completed_at IS NULL
AND created_by = $1
ORDER BY priority DESC, created_at ASC;

-- name: GetNoCompletedTasks :one
SELECT count(*) AS total
FROM task
WHERE completed_at >= $1
AND completed_at < $2
AND created_by = $3;

-- name: CompleteTasks :one
UPDATE task
SET completed_at = NOW()
WHERE id = ANY($1::int[])
AND created_by = $2
AND completed_at IS NULL
RETURNING count(*) AS total;

-- name: MarkTaskAsDone :exec
UPDATE task
SET completed_at = NOW()
WHERE id = $1
AND created_by = $2;
