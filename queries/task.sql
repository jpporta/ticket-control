-- name: TotalTasksFromUserToday :one
SELECT COUNT(*) AS total FROM task 
WHERE created_by = $1 
AND DATE(created_at) = CURRENT_DATE;


-- name: CreateTask :one
INSERT INTO task (title, description, priority, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id;
