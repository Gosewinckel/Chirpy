-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
	GEN_RANDOM_UUID(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING *;

-- name: ClearUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;
