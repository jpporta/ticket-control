-- name: TotalListsFromUser :one
SELECT count(*) AS total FROM list
WHERE created_by = $1
AND created_at >= $2
AND created_at < $3;

-- name: CreateList :one
INSERT INTO list (title, content, created_by)
VALUES ($1, $2, $3)
RETURNING id;

-- name: DeleteLastList :exec
DELETE FROM list
WHERE id = (
	SELECT id FROM task as t
	WHERE t.created_by = $1
	ORDER BY created_at DESC
	LIMIT 1
);
