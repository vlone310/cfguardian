-- Drop config_schemas table and its indexes
DROP INDEX IF EXISTS idx_config_schemas_created_by_user_id;
DROP INDEX IF EXISTS idx_config_schemas_name;
DROP TABLE IF EXISTS config_schemas CASCADE;

