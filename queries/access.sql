-- name: AddAccess :exec
INSERT INTO access (
	user_id,
  ip_address
) VALUES ($1, $2);

