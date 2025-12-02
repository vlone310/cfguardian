-- name: CreateConfig :one
INSERT INTO configs (
    project_id,
    key,
    schema_id,
    version,
    content,
    updated_by_user_id
) VALUES (
    $1, $2, $3, 1, $4, $5
)
RETURNING *;

-- name: GetConfig :one
SELECT * FROM configs
WHERE project_id = $1 AND key = $2
LIMIT 1;

-- name: GetConfigWithVersion :one
SELECT * FROM configs
WHERE project_id = $1 AND key = $2 AND version = $3
LIMIT 1;

-- name: ListConfigsByProject :many
SELECT * FROM configs
WHERE project_id = $1
ORDER BY key;

-- name: ListConfigsBySchema :many
SELECT * FROM configs
WHERE schema_id = $1
ORDER BY project_id, key;

-- name: UpdateConfig :one
UPDATE configs
SET
    content = $4,
    version = version + 1,
    updated_by_user_id = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE project_id = $1 AND key = $2 AND version = $3
RETURNING *;

-- name: ChangeConfigSchema :one
UPDATE configs
SET
    schema_id = $3,
    updated_by_user_id = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE project_id = $1 AND key = $2
RETURNING *;

-- name: DeleteConfig :exec
DELETE FROM configs
WHERE project_id = $1 AND key = $2;

-- name: ConfigExists :one
SELECT EXISTS(
    SELECT 1 FROM configs
    WHERE project_id = $1 AND key = $2
) AS exists;

-- name: GetConfigVersion :one
SELECT version FROM configs
WHERE project_id = $1 AND key = $2
LIMIT 1;

-- name: CountConfigsByProject :one
SELECT COUNT(*) FROM configs
WHERE project_id = $1;

-- name: CountConfigsBySchema :one
SELECT COUNT(*) FROM configs
WHERE schema_id = $1;

-- name: SearchConfigsByKey :many
SELECT * FROM configs
WHERE project_id = $1 AND key ILIKE $2
ORDER BY key
LIMIT $3;

-- name: GetConfigsUpdatedAfter :many
SELECT * FROM configs
WHERE project_id = $1 AND updated_at > $2
ORDER BY updated_at DESC;

-- name: GetConfigsUpdatedByUser :many
SELECT * FROM configs
WHERE updated_by_user_id = $1
ORDER BY updated_at DESC
LIMIT $2;

-- Optimistic locking helper - get current version for update
-- name: LockConfigForUpdate :one
SELECT version FROM configs
WHERE project_id = $1 AND key = $2
FOR UPDATE;

