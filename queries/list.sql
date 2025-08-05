-- name: TotalListsFromUserToday :one
SELECT COUNT(*) AS total FROM list 
WHERE created_by = $1 
AND DATE(created_at) = CURRENT_DATE;

-- name: CreateList :one
INSERT INTO list (title, content, created_by)
VALUES ($1, $2, $3)
RETURNING id;
