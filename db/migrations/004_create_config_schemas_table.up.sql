-- Create config_schemas table for reusable validation structures
CREATE TABLE IF NOT EXISTS config_schemas (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    schema_content TEXT NOT NULL,
    created_by_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_config_schemas_created_by_user
        FOREIGN KEY (created_by_user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_config_schemas_name ON config_schemas(name);
CREATE INDEX IF NOT EXISTS idx_config_schemas_created_by_user_id ON config_schemas(created_by_user_id);

-- Add comments
COMMENT ON TABLE config_schemas IS 'Reusable JSON Schema definitions for configuration validation';
COMMENT ON COLUMN config_schemas.id IS 'Unique schema identifier (UUID)';
COMMENT ON COLUMN config_schemas.name IS 'Schema name for reference';
COMMENT ON COLUMN config_schemas.schema_content IS 'JSON Schema definition';
COMMENT ON COLUMN config_schemas.created_by_user_id IS 'User who created this schema';

