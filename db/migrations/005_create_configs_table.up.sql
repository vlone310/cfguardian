-- Create configs table for current authoritative state
-- This table requires Raft/etcd consistency for strong CP guarantees
CREATE TABLE IF NOT EXISTS configs (
    project_id VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    schema_id VARCHAR(255) NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    content JSONB NOT NULL,
    updated_by_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, key),
    CONSTRAINT fk_configs_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_configs_schema
        FOREIGN KEY (schema_id)
        REFERENCES config_schemas(id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_configs_updated_by_user
        FOREIGN KEY (updated_by_user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT chk_configs_version_positive
        CHECK (version > 0)
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_configs_project_id ON configs(project_id);
CREATE INDEX IF NOT EXISTS idx_configs_key ON configs(key);
CREATE INDEX IF NOT EXISTS idx_configs_schema_id ON configs(schema_id);
CREATE INDEX IF NOT EXISTS idx_configs_updated_by_user_id ON configs(updated_by_user_id);
CREATE INDEX IF NOT EXISTS idx_configs_updated_at ON configs(updated_at);

-- Create GIN index for JSONB content for faster JSON queries
CREATE INDEX IF NOT EXISTS idx_configs_content_gin ON configs USING GIN (content);

-- Add comments
COMMENT ON TABLE configs IS 'Current authoritative configuration state (requires Raft consensus)';
COMMENT ON COLUMN configs.project_id IS 'Reference to project (part of composite PK)';
COMMENT ON COLUMN configs.key IS 'Configuration key (part of composite PK)';
COMMENT ON COLUMN configs.schema_id IS 'Reference to validation schema';
COMMENT ON COLUMN configs.version IS 'Optimistic locking version counter';
COMMENT ON COLUMN configs.content IS 'Canonical configuration data (JSON)';
COMMENT ON COLUMN configs.updated_by_user_id IS 'User who last updated this config';

