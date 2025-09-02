-- name: AddAccess :exec
INSERT INTO access (
	user_id,
  ip_address,
	path,
  method
) VALUES ($1, $2, $3, $4);

-- name: GetAccessStats :one
SELECT count(*) AS total FROM access
WHERE user_id = $1
AND path = $2
AND method = $3
AND accessed_at >= $4
AND accessed_at < $5;

