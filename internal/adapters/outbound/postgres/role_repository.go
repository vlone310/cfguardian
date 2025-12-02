package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlone310/cfguardian/internal/adapters/outbound/postgres/sqlc"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// RoleRepositoryAdapter implements outbound.RoleRepository using PostgreSQL
type RoleRepositoryAdapter struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewRoleRepositoryAdapter creates a new PostgreSQL role repository
func NewRoleRepositoryAdapter(pool *pgxpool.Pool) *RoleRepositoryAdapter {
	return &RoleRepositoryAdapter{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Assign assigns or updates a role for a user in a project (upsert)
func (r *RoleRepositoryAdapter) Assign(ctx context.Context, params outbound.AssignRoleParams) (*outbound.Role, error) {
	role, err := r.queries.AssignRole(ctx, params.UserID, params.ProjectID, sqlc.RoleLevel(params.RoleLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}
	
	return r.modelToOutbound(&role), nil
}

// Get retrieves a specific role
func (r *RoleRepositoryAdapter) Get(ctx context.Context, userID, projectID string) (*outbound.Role, error) {
	role, err := r.queries.GetRole(ctx, userID, projectID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	
	return r.modelToOutbound(&role), nil
}

// GetUserRole retrieves just the role level for a user in a project
func (r *RoleRepositoryAdapter) GetUserRole(ctx context.Context, userID, projectID string) (outbound.RoleLevel, error) {
	roleLevel, err := r.queries.GetUserRole(ctx, userID, projectID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("role not found")
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	
	return outbound.RoleLevel(roleLevel), nil
}

// ListUserRoles retrieves all roles for a specific user (with project names)
func (r *RoleRepositoryAdapter) ListUserRoles(ctx context.Context, userID string) ([]*outbound.UserRoleWithProject, error) {
	roles, err := r.queries.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}
	
	result := make([]*outbound.UserRoleWithProject, len(roles))
	for i, role := range roles {
		result[i] = &outbound.UserRoleWithProject{
			Role: outbound.Role{
				UserID:    role.UserID,
				ProjectID: role.ProjectID,
				RoleLevel: outbound.RoleLevel(role.RoleLevel),
				CreatedAt: role.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt: role.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			},
			ProjectName: role.ProjectName,
		}
	}
	
	return result, nil
}

// ListProjectRoles retrieves all roles for a specific project (with user emails)
func (r *RoleRepositoryAdapter) ListProjectRoles(ctx context.Context, projectID string) ([]*outbound.ProjectRoleWithUser, error) {
	roles, err := r.queries.ListProjectRoles(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project roles: %w", err)
	}
	
	result := make([]*outbound.ProjectRoleWithUser, len(roles))
	for i, role := range roles {
		result[i] = &outbound.ProjectRoleWithUser{
			Role: outbound.Role{
				UserID:    role.UserID,
				ProjectID: role.ProjectID,
				RoleLevel: outbound.RoleLevel(role.RoleLevel),
				CreatedAt: role.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt: role.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			},
			UserEmail: role.UserEmail,
		}
	}
	
	return result, nil
}

// ListByLevel retrieves all roles with a specific level
func (r *RoleRepositoryAdapter) ListByLevel(ctx context.Context, level outbound.RoleLevel) ([]*outbound.Role, error) {
	roleRows, err := r.queries.ListRolesByLevel(ctx, sqlc.RoleLevel(level))
	if err != nil {
		return nil, fmt.Errorf("failed to list roles by level: %w", err)
	}
	
	result := make([]*outbound.Role, len(roleRows))
	for i, row := range roleRows {
		result[i] = &outbound.Role{
			UserID:    row.UserID,
			ProjectID: row.ProjectID,
			RoleLevel: outbound.RoleLevel(row.RoleLevel),
			CreatedAt: row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: row.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	
	return result, nil
}

// Update updates a role
func (r *RoleRepositoryAdapter) Update(ctx context.Context, userID, projectID string, level outbound.RoleLevel) (*outbound.Role, error) {
	role, err := r.queries.UpdateRole(ctx, userID, projectID, sqlc.RoleLevel(level))
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	
	return r.modelToOutbound(&role), nil
}

// Revoke revokes a specific role
func (r *RoleRepositoryAdapter) Revoke(ctx context.Context, userID, projectID string) error {
	err := r.queries.RevokeRole(ctx, userID, projectID)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}
	return nil
}

// RevokeAllUserRoles revokes all roles for a user
func (r *RoleRepositoryAdapter) RevokeAllUserRoles(ctx context.Context, userID string) error {
	err := r.queries.RevokeAllUserRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user roles: %w", err)
	}
	return nil
}

// RevokeAllProjectRoles revokes all roles for a project
func (r *RoleRepositoryAdapter) RevokeAllProjectRoles(ctx context.Context, projectID string) error {
	err := r.queries.RevokeAllProjectRoles(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to revoke all project roles: %w", err)
	}
	return nil
}

// Exists checks if a role exists
func (r *RoleRepositoryAdapter) Exists(ctx context.Context, userID, projectID string) (bool, error) {
	exists, err := r.queries.RoleExists(ctx, userID, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}
	return exists, nil
}

// HasRole checks if a user has a specific role in a project
func (r *RoleRepositoryAdapter) HasRole(ctx context.Context, userID, projectID string, level outbound.RoleLevel) (bool, error) {
	roleLevel, err := r.GetUserRole(ctx, userID, projectID)
	if err != nil {
		// If role not found, user doesn't have the role
		return false, nil
	}
	
	return roleLevel == level, nil
}

// HasMinimumRole checks if a user has at least the specified role level
func (r *RoleRepositoryAdapter) HasMinimumRole(ctx context.Context, userID, projectID string, minLevel outbound.RoleLevel) (bool, error) {
	roleLevel, err := r.GetUserRole(ctx, userID, projectID)
	if err != nil {
		// If role not found, user doesn't have minimum role
		return false, nil
	}
	
	// Check hierarchical permissions: admin > editor > viewer
	switch minLevel {
	case outbound.RoleLevelViewer:
		return roleLevel == outbound.RoleLevelViewer ||
			roleLevel == outbound.RoleLevelEditor ||
			roleLevel == outbound.RoleLevelAdmin, nil
	case outbound.RoleLevelEditor:
		return roleLevel == outbound.RoleLevelEditor ||
			roleLevel == outbound.RoleLevelAdmin, nil
	case outbound.RoleLevelAdmin:
		return roleLevel == outbound.RoleLevelAdmin, nil
	default:
		return false, nil
	}
}

// CountByProject returns the number of roles in a project
func (r *RoleRepositoryAdapter) CountByProject(ctx context.Context, projectID string) (int64, error) {
	count, err := r.queries.CountRolesByProject(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles by project: %w", err)
	}
	return count, nil
}

// CountByUser returns the number of roles for a user
func (r *RoleRepositoryAdapter) CountByUser(ctx context.Context, userID string) (int64, error) {
	count, err := r.queries.CountRolesByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles by user: %w", err)
	}
	return count, nil
}

// modelToOutbound converts SQLC model to outbound model
func (r *RoleRepositoryAdapter) modelToOutbound(role *sqlc.Role) *outbound.Role {
	return &outbound.Role{
		UserID:    role.UserID,
		ProjectID: role.ProjectID,
		RoleLevel: outbound.RoleLevel(role.RoleLevel),
		CreatedAt: role.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: role.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

