-- name: TotalLinksFromUser :one
SELECT count(*) AS total FROM link
WHERE created_by = $1
AND created_at >= $2
AND created_at < $3;

-- name: CreateLink :one
INSERT INTO link (url, title, created_by)
VALUES ($1, $2, $3)
RETURNING id;

-- name: DeleteLastLink :exec
DELETE FROM link
WHERE id = (
	SELECT id FROM link as t
	WHERE t.created_by = $1
	ORDER BY created_at DESC
	LIMIT 1
);

-- name: GetLinkByID :one
SELECT url FROM link WHERE id = $1;
