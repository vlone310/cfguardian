-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    password_hash
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET
    email = COALESCE(sqlc.narg('email'), email),
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UserExists :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE id = $1
) AS exists;

-- name: UserExistsByEmail :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
) AS exists;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

