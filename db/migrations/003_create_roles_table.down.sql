-- Drop roles table and its indexes
DROP INDEX IF EXISTS idx_roles_role_level;
DROP INDEX IF EXISTS idx_roles_project_id;
DROP INDEX IF EXISTS idx_roles_user_id;
DROP TABLE IF EXISTS roles CASCADE;

-- Drop role_level enum type
DROP TYPE IF EXISTS role_level;

