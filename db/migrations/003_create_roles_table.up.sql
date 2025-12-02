-- Create role_level enum type
CREATE TYPE role_level AS ENUM ('admin', 'editor', 'viewer');

-- Create roles table for access control
CREATE TABLE IF NOT EXISTS roles (
    user_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    role_level role_level NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, project_id),
    CONSTRAINT fk_roles_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_roles_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_roles_user_id ON roles(user_id);
CREATE INDEX IF NOT EXISTS idx_roles_project_id ON roles(project_id);
CREATE INDEX IF NOT EXISTS idx_roles_role_level ON roles(role_level);

-- Add comments
COMMENT ON TABLE roles IS 'User roles for access control within projects';
COMMENT ON COLUMN roles.user_id IS 'Reference to user';
COMMENT ON COLUMN roles.project_id IS 'Reference to project';
COMMENT ON COLUMN roles.role_level IS 'Role level: admin, editor, or viewer';

