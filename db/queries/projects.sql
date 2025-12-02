-- name: CreateProject :one
INSERT INTO projects (
    id,
    name,
    api_key,
    owner_user_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1
LIMIT 1;

-- name: GetProjectByAPIKey :one
SELECT * FROM projects
WHERE api_key = $1
LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects
ORDER BY created_at DESC;

-- name: ListProjectsByOwner :many
SELECT * FROM projects
WHERE owner_user_id = $1
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(sqlc.narg('name'), name),
    api_key = COALESCE(sqlc.narg('api_key'), api_key),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;

-- name: ProjectExists :one
SELECT EXISTS(
    SELECT 1 FROM projects WHERE id = $1
) AS exists;

-- name: ProjectExistsByAPIKey :one
SELECT EXISTS(
    SELECT 1 FROM projects WHERE api_key = $1
) AS exists;

-- name: ProjectExistsByName :one
SELECT EXISTS(
    SELECT 1 FROM projects WHERE name = $1
) AS exists;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects;

-- name: CountProjectsByOwner :one
SELECT COUNT(*) FROM projects
WHERE owner_user_id = $1;

