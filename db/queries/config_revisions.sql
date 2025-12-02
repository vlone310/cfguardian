-- name: CreateConfigRevision :one
INSERT INTO config_revisions (
    id,
    project_id,
    config_key,
    version,
    content,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetConfigRevision :one
SELECT * FROM config_revisions
WHERE id = $1
LIMIT 1;

-- name: GetConfigRevisionByVersion :one
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2 AND version = $3
LIMIT 1;

-- name: ListConfigRevisions :many
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2
ORDER BY version DESC;

-- name: ListConfigRevisionsPaginated :many
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2
ORDER BY version DESC
LIMIT $3 OFFSET $4;

-- name: ListAllRevisionsByProject :many
SELECT * FROM config_revisions
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: ListRevisionsByUser :many
SELECT * FROM config_revisions
WHERE created_by_user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetLatestRevision :one
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2
ORDER BY version DESC
LIMIT 1;

-- name: GetLatestNRevisions :many
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2
ORDER BY version DESC
LIMIT $3;

-- name: CountRevisions :one
SELECT COUNT(*) FROM config_revisions
WHERE project_id = $1 AND config_key = $2;

-- name: CountRevisionsByProject :one
SELECT COUNT(*) FROM config_revisions
WHERE project_id = $1;

-- name: GetRevisionsCreatedAfter :many
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2 AND created_at > $3
ORDER BY version ASC;

-- name: GetRevisionsInVersionRange :many
SELECT * FROM config_revisions
WHERE project_id = $1 AND config_key = $2
  AND version >= $3 AND version <= $4
ORDER BY version ASC;

-- name: DeleteOldRevisions :exec
DELETE FROM config_revisions
WHERE project_id = $1 AND config_key = $2
  AND version < $3;

-- name: GetRevisionHistory :many
SELECT 
    cr.*,
    u.email as created_by_email
FROM config_revisions cr
JOIN users u ON cr.created_by_user_id = u.id
WHERE cr.project_id = $1 AND cr.config_key = $2
ORDER BY cr.version DESC
LIMIT $3;

