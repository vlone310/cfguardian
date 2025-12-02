-- Drop projects table and its indexes
DROP INDEX IF EXISTS idx_projects_name;
DROP INDEX IF EXISTS idx_projects_owner_user_id;
DROP INDEX IF EXISTS idx_projects_api_key;
DROP TABLE IF EXISTS projects CASCADE;

