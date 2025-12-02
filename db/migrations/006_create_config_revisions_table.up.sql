-- Create config_revisions table for immutable audit log
CREATE TABLE IF NOT EXISTS config_revisions (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    config_key VARCHAR(255) NOT NULL,
    version BIGINT NOT NULL,
    content JSONB NOT NULL,
    created_by_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_config_revisions_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_config_revisions_config
        FOREIGN KEY (project_id, config_key)
        REFERENCES configs(project_id, key)
        ON DELETE CASCADE,
    CONSTRAINT fk_config_revisions_created_by_user
        FOREIGN KEY (created_by_user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT chk_config_revisions_version_positive
        CHECK (version > 0),
    CONSTRAINT uq_config_revisions_project_key_version
        UNIQUE (project_id, config_key, version)
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_config_revisions_project_id ON config_revisions(project_id);
CREATE INDEX IF NOT EXISTS idx_config_revisions_config_key ON config_revisions(config_key);
CREATE INDEX IF NOT EXISTS idx_config_revisions_project_key ON config_revisions(project_id, config_key);
CREATE INDEX IF NOT EXISTS idx_config_revisions_version ON config_revisions(version);
CREATE INDEX IF NOT EXISTS idx_config_revisions_created_at ON config_revisions(created_at);
CREATE INDEX IF NOT EXISTS idx_config_revisions_created_by_user_id ON config_revisions(created_by_user_id);

-- Create GIN index for JSONB content for faster JSON queries
CREATE INDEX IF NOT EXISTS idx_config_revisions_content_gin ON config_revisions USING GIN (content);

-- Add comments
COMMENT ON TABLE config_revisions IS 'Immutable audit log of all configuration changes';
COMMENT ON COLUMN config_revisions.id IS 'Unique revision identifier (UUID)';
COMMENT ON COLUMN config_revisions.project_id IS 'Reference to project';
COMMENT ON COLUMN config_revisions.config_key IS 'Configuration key';
COMMENT ON COLUMN config_revisions.version IS 'Version number of this revision';
COMMENT ON COLUMN config_revisions.content IS 'Full historical configuration data (JSON)';
COMMENT ON COLUMN config_revisions.created_by_user_id IS 'User who created this revision';

