-- name: AssignRole :one
INSERT INTO roles (
    user_id,
    project_id,
    role_level
) VALUES (
    $1, $2, $3
)
ON CONFLICT (user_id, project_id)
DO UPDATE SET
    role_level = EXCLUDED.role_level,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE user_id = $1 AND project_id = $2
LIMIT 1;

-- name: GetUserRole :one
SELECT role_level FROM roles
WHERE user_id = $1 AND project_id = $2
LIMIT 1;

-- name: ListUserRoles :many
SELECT r.*, p.name as project_name
FROM roles r
JOIN projects p ON r.project_id = p.id
WHERE r.user_id = $1
ORDER BY p.name;

-- name: ListProjectRoles :many
SELECT r.*, u.email as user_email
FROM roles r
JOIN users u ON r.user_id = u.id
WHERE r.project_id = $1
ORDER BY r.role_level, u.email;

-- name: ListRolesByLevel :many
SELECT r.*, u.email as user_email, p.name as project_name
FROM roles r
JOIN users u ON r.user_id = u.id
JOIN projects p ON r.project_id = p.id
WHERE r.role_level = $1
ORDER BY p.name, u.email;

-- name: UpdateRole :one
UPDATE roles
SET
    role_level = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND project_id = $2
RETURNING *;

-- name: RevokeRole :exec
DELETE FROM roles
WHERE user_id = $1 AND project_id = $2;

-- name: RevokeAllUserRoles :exec
DELETE FROM roles
WHERE user_id = $1;

-- name: RevokeAllProjectRoles :exec
DELETE FROM roles
WHERE project_id = $1;

-- name: RoleExists :one
SELECT EXISTS(
    SELECT 1 FROM roles
    WHERE user_id = $1 AND project_id = $2
) AS exists;

-- name: UserHasRole :one
SELECT EXISTS(
    SELECT 1 FROM roles
    WHERE user_id = $1 AND project_id = $2 AND role_level = $3
) AS exists;

-- name: UserHasMinimumRole :one
SELECT EXISTS(
    SELECT 1 FROM roles
    WHERE user_id = $1 AND project_id = $2
    AND (
        ($3 = 'viewer' AND role_level IN ('viewer', 'editor', 'admin')) OR
        ($3 = 'editor' AND role_level IN ('editor', 'admin')) OR
        ($3 = 'admin' AND role_level = 'admin')
    )
) AS exists;

-- name: CountRolesByProject :one
SELECT COUNT(*) FROM roles
WHERE project_id = $1;

-- name: CountRolesByUser :one
SELECT COUNT(*) FROM roles
WHERE user_id = $1;

