-- name: CreateConfigSchema :one
INSERT INTO config_schemas (
    id,
    name,
    schema_content,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetConfigSchemaByID :one
SELECT * FROM config_schemas
WHERE id = $1
LIMIT 1;

-- name: GetConfigSchemaByName :one
SELECT * FROM config_schemas
WHERE name = $1
LIMIT 1;

-- name: ListConfigSchemas :many
SELECT * FROM config_schemas
ORDER BY name;

-- name: ListConfigSchemasByCreator :many
SELECT * FROM config_schemas
WHERE created_by_user_id = $1
ORDER BY created_at DESC;

-- name: UpdateConfigSchema :one
UPDATE config_schemas
SET
    name = COALESCE(sqlc.narg('name'), name),
    schema_content = COALESCE(sqlc.narg('schema_content'), schema_content),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteConfigSchema :exec
DELETE FROM config_schemas
WHERE id = $1;

-- name: ConfigSchemaExists :one
SELECT EXISTS(
    SELECT 1 FROM config_schemas WHERE id = $1
) AS exists;

-- name: ConfigSchemaExistsByName :one
SELECT EXISTS(
    SELECT 1 FROM config_schemas WHERE name = $1
) AS exists;

-- name: CountConfigSchemas :one
SELECT COUNT(*) FROM config_schemas;

-- name: CountConfigsUsingSchema :one
SELECT COUNT(*) FROM configs
WHERE schema_id = $1;

