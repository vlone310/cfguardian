-- Drop config_revisions table and its indexes
DROP INDEX IF EXISTS idx_config_revisions_content_gin;
DROP INDEX IF EXISTS idx_config_revisions_created_by_user_id;
DROP INDEX IF EXISTS idx_config_revisions_created_at;
DROP INDEX IF EXISTS idx_config_revisions_version;
DROP INDEX IF EXISTS idx_config_revisions_project_key;
DROP INDEX IF EXISTS idx_config_revisions_config_key;
DROP INDEX IF EXISTS idx_config_revisions_project_id;
DROP TABLE IF EXISTS config_revisions CASCADE;

