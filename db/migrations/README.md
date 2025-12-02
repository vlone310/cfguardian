# Database Migrations

This directory contains database migration files for GoConfig Guardian.

## Migration Files

Migrations are numbered sequentially and include both `up` and `down` versions:

| # | Name | Description |
|---|------|-------------|
| 001 | `create_users_table` | Creates users table for management UI access |
| 002 | `create_projects_table` | Creates projects table for multi-tenancy |
| 003 | `create_roles_table` | Creates roles table and role_level enum for RBAC |
| 004 | `create_config_schemas_table` | Creates config_schemas table for JSON Schema validation |
| 005 | `create_configs_table` | Creates configs table (Raft-backed, optimistic locking) |
| 006 | `create_config_revisions_table` | Creates config_revisions table for audit log |

## Database Schema

### Entity Relationship Diagram

```
┌─────────────┐
│   users     │
└──────┬──────┘
       │
       │ owner_user_id
       ▼
┌─────────────┐        ┌──────────────────┐
│  projects   │◄───────┤  config_schemas  │
└──────┬──────┘        └──────────────────┘
       │                        │
       │ project_id             │ schema_id
       ▼                        ▼
┌─────────────┐        ┌──────────────┐
│   roles     │        │   configs    │ ⭐ Raft-backed
└─────────────┘        └──────┬───────┘
                              │
                              │ (project_id, config_key)
                              ▼
                       ┌──────────────────┐
                       │ config_revisions │
                       └──────────────────┘
```

### Tables

#### 1. users
Management UI users with authentication.
- **PK**: `id` (VARCHAR)
- **Unique**: `email`
- Stores: email, password_hash

#### 2. projects
Multi-tenancy scoping with API keys.
- **PK**: `id` (VARCHAR)
- **Unique**: `api_key`
- **FK**: `owner_user_id` → users(id)
- Stores: name, api_key

#### 3. roles
Access control (RBAC) for users within projects.
- **Composite PK**: (`user_id`, `project_id`)
- **FK**: `user_id` → users(id)
- **FK**: `project_id` → projects(id)
- **Enum**: `role_level` (admin, editor, viewer)

#### 4. config_schemas
Reusable JSON Schema definitions.
- **PK**: `id` (VARCHAR)
- **FK**: `created_by_user_id` → users(id)
- Stores: name, schema_content (TEXT)

#### 5. configs ⭐ (Raft-backed)
Current authoritative configuration state with strong consistency.
- **Composite PK**: (`project_id`, `key`)
- **FK**: `project_id` → projects(id)
- **FK**: `schema_id` → config_schemas(id)
- **FK**: `updated_by_user_id` → users(id)
- **Critical**: `version` BIGINT for optimistic locking
- **Storage**: `content` (JSONB) for configuration data
- **Consistency**: Requires Raft consensus

#### 6. config_revisions
Immutable audit log of all configuration changes.
- **PK**: `id` (VARCHAR)
- **FK**: `project_id` → projects(id)
- **FK**: (`project_id`, `config_key`) → configs(project_id, key)
- **FK**: `created_by_user_id` → users(id)
- **Unique**: (`project_id`, `config_key`, `version`)
- Stores: version, content (JSONB)

## Running Migrations

### Using Make Commands

```bash
# Apply all pending migrations
make migrate-up

# Apply one migration
make migrate-up-one

# Rollback last migration
make migrate-down

# Show current version
make migrate-version

# Create new migration
make migrate-create NAME=add_new_feature

# Force specific version (use with caution)
make migrate-force VERSION=5
```

### Using Migration Script Directly

```bash
# Apply all migrations
./scripts/migrate.sh up

# Apply N migrations
./scripts/migrate.sh up 2

# Rollback one migration
./scripts/migrate.sh down 1

# Check current version
./scripts/migrate.sh version

# Create new migration
./scripts/migrate.sh create add_new_feature
```

## Important Notes

### Optimistic Locking
The `configs` table uses a `version` field for optimistic locking:
- Version starts at 1
- Increments on every update
- Updates fail with 409 Conflict if version mismatches
- Prevents concurrent modification conflicts

### Raft Consistency
The `configs` table is the **only table requiring Raft consensus**:
- All writes to `configs` must go through the Raft leader
- Ensures all nodes agree on the latest configuration version
- Provides strong consistency (CP in CAP theorem)
- Other tables use standard PostgreSQL ACID guarantees

### Indexes
All tables have appropriate indexes for:
- Primary key lookups
- Foreign key relationships
- Common query patterns
- JSONB content (GIN indexes for configs and revisions)

### Cascade Behavior
- Deleting a user cascades to their projects, roles, and schemas
- Deleting a project cascades to roles, configs, and revisions
- Deleting a config schema is RESTRICTED (must have no configs using it)

## Migration Best Practices

1. **Always test migrations locally first**
2. **Review both up and down migrations**
3. **Never modify existing migration files** - create new ones instead
4. **Backup production data before migrating**
5. **Test rollback (down) migrations** before applying to production
6. **Version migrations sequentially** - don't skip numbers

