-- name: CreateUser :one
INSERT INTO public."user" (
		name,
		api_key
) VALUES ($1, $2)
RETURNING id;

-- name: GetUserByKey :one
select id from public."user" where api_key = $1;
