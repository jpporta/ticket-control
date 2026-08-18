-- name: TotalLettersFromUser :one
SELECT count(*) AS total FROM letter
WHERE created_by = $1
AND created_at >= $2
AND created_at < $3;

-- name: CreateLetter :one
INSERT INTO letter (title, content, recipient, sender, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: DeleteLastLetter :exec
DELETE FROM letter
WHERE id = (
	SELECT id FROM letter as l
	WHERE l.created_by = $1
	ORDER BY created_at DESC
	LIMIT 1
);

-- name: GetLetterByID :one
SELECT * FROM letter WHERE id = $1;
