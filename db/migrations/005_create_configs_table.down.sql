-- Drop configs table and its indexes
DROP INDEX IF EXISTS idx_configs_content_gin;
DROP INDEX IF EXISTS idx_configs_updated_at;
DROP INDEX IF EXISTS idx_configs_updated_by_user_id;
DROP INDEX IF EXISTS idx_configs_schema_id;
DROP INDEX IF EXISTS idx_configs_key;
DROP INDEX IF EXISTS idx_configs_project_id;
DROP TABLE IF EXISTS configs CASCADE;

