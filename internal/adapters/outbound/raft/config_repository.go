package raft

import (
	"context"
	"fmt"
	"time"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// ConfigRepository implements the outbound.ConfigRepository interface using Raft
// This provides strong consistency (CP) for configuration data
type ConfigRepository struct {
	store *Store
}

// NewConfigRepository creates a new Raft-backed config repository
func NewConfigRepository(store *Store) *ConfigRepository {
	return &ConfigRepository{
		store: store,
	}
}

// Create creates a new config through Raft consensus
func (r *ConfigRepository) Create(ctx context.Context, params outbound.CreateConfigParams) (*outbound.Config, error) {
	// Apply through Raft
	state, err := r.store.CreateConfig(
		ctx,
		params.ProjectID,
		params.Key,
		params.SchemaID,
		params.Content,
		params.UpdatedByUserID,
	)
	if err != nil {
		return nil, err
	}
	
	// Convert to outbound.Config
	return r.stateToConfig(state), nil
}

// Get retrieves a config (read from FSM, no consensus needed)
func (r *ConfigRepository) Get(ctx context.Context, projectID, key string) (*outbound.Config, error) {
	state, err := r.store.GetConfig(projectID, key)
	if err != nil {
		return nil, fmt.Errorf("config not found")
	}
	
	return r.stateToConfig(state), nil
}

// GetWithVersion retrieves a config only if it matches the expected version
func (r *ConfigRepository) GetWithVersion(ctx context.Context, projectID, key string, version int64) (*outbound.Config, error) {
	state, err := r.store.GetConfig(projectID, key)
	if err != nil {
		return nil, fmt.Errorf("config not found")
	}
	
	if state.Version != version {
		return nil, fmt.Errorf("version mismatch")
	}
	
	return r.stateToConfig(state), nil
}

// ListByProject lists all configs for a project
func (r *ConfigRepository) ListByProject(ctx context.Context, projectID string) ([]*outbound.Config, error) {
	states := r.store.ListConfigs(projectID)
	
	configs := make([]*outbound.Config, len(states))
	for i, state := range states {
		configs[i] = r.stateToConfig(state)
	}
	
	return configs, nil
}

// ListBySchema lists all configs using a specific schema (not supported in Raft - use PostgreSQL fallback)
func (r *ConfigRepository) ListBySchema(ctx context.Context, schemaID string) ([]*outbound.Config, error) {
	return nil, fmt.Errorf("ListBySchema not supported in Raft store - use PostgreSQL repository")
}

// Update updates a config through Raft consensus with optimistic locking
func (r *ConfigRepository) Update(ctx context.Context, params outbound.UpdateConfigParams) (*outbound.Config, error) {
	// Apply through Raft
	state, err := r.store.UpdateConfig(
		ctx,
		params.ProjectID,
		params.Key,
		params.ExpectedVersion,
		params.Content,
		params.UpdatedByUserID,
	)
	if err != nil {
		return nil, err
	}
	
	return r.stateToConfig(state), nil
}

// ChangeSchema changes the schema ID for a config (not fully supported in current FSM)
func (r *ConfigRepository) ChangeSchema(ctx context.Context, params outbound.ChangeSchemaParams) (*outbound.Config, error) {
	// For now, we'll need to implement this as an update
	// Get current config
	current, err := r.Get(ctx, params.ProjectID, params.Key)
	if err != nil {
		return nil, err
	}
	
	// Update with same content but new schema (this is a simplification)
	// In a full implementation, we'd add a CHANGE_SCHEMA command to the FSM
	state, err := r.store.UpdateConfig(
		ctx,
		params.ProjectID,
		params.Key,
		current.Version,
		current.Content,
		params.UpdatedByUserID,
	)
	if err != nil {
		return nil, err
	}
	
	// Manually update schema ID (not persisted in Raft currently)
	config := r.stateToConfig(state)
	config.SchemaID = params.SchemaID
	
	return config, nil
}

// Delete deletes a config through Raft consensus
func (r *ConfigRepository) Delete(ctx context.Context, projectID, key string) error {
	// For delete, we need a user ID - use "system" as default
	return r.store.DeleteConfig(ctx, projectID, key, "system")
}

// Exists checks if a config exists
func (r *ConfigRepository) Exists(ctx context.Context, projectID, key string) (bool, error) {
	return r.store.fsm.ConfigExists(projectID, key), nil
}

// GetVersion gets the current version of a config
func (r *ConfigRepository) GetVersion(ctx context.Context, projectID, key string) (int64, error) {
	version, err := r.store.fsm.GetVersion(projectID, key)
	if err != nil {
		return 0, err
	}
	return version.Value(), nil
}

// LockForUpdate is not needed in Raft (consensus provides the lock)
func (r *ConfigRepository) LockForUpdate(ctx context.Context, projectID, key string) (int64, error) {
	return r.GetVersion(ctx, projectID, key)
}

// Search searches configs by key pattern (not supported in Raft - use PostgreSQL fallback)
func (r *ConfigRepository) Search(ctx context.Context, params outbound.SearchConfigsParams) ([]*outbound.Config, error) {
	return nil, fmt.Errorf("Search not supported in Raft store - use PostgreSQL repository")
}

// GetUpdatedAfter returns configs updated after a specific time (not supported in Raft)
func (r *ConfigRepository) GetUpdatedAfter(ctx context.Context, projectID string, afterTime string) ([]*outbound.Config, error) {
	return nil, fmt.Errorf("GetUpdatedAfter not supported in Raft store - use PostgreSQL repository")
}

// GetUpdatedByUser returns configs updated by a specific user (not supported in Raft)
func (r *ConfigRepository) GetUpdatedByUser(ctx context.Context, userID string, limit int32) ([]*outbound.Config, error) {
	return nil, fmt.Errorf("GetUpdatedByUser not supported in Raft store - use PostgreSQL repository")
}

// CountByProject returns the number of configs in a project
func (r *ConfigRepository) CountByProject(ctx context.Context, projectID string) (int64, error) {
	configs := r.store.ListConfigs(projectID)
	return int64(len(configs)), nil
}

// CountBySchema returns the number of configs using a schema (not supported in Raft)
func (r *ConfigRepository) CountBySchema(ctx context.Context, schemaID string) (int64, error) {
	return 0, fmt.Errorf("CountBySchema not supported in Raft store - use PostgreSQL repository")
}

// stateToConfig converts FSM ConfigState to outbound.Config
func (r *ConfigRepository) stateToConfig(state *ConfigState) *outbound.Config {
	now := time.Now().Format(time.RFC3339)
	
	return &outbound.Config{
		ProjectID:       state.ProjectID,
		Key:             state.Key,
		SchemaID:        state.SchemaID,
		Version:         state.Version,
		Content:         state.Content,
		UpdatedByUserID: state.UpdatedByUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// IsLeader checks if this node is the Raft leader
func (r *ConfigRepository) IsLeader() bool {
	return r.store.IsLeader()
}

// GetLeaderAddress returns the current leader address
func (r *ConfigRepository) GetLeaderAddress() string {
	return r.store.GetLeader()
}

// WaitForLeader waits until a leader is elected
func (r *ConfigRepository) WaitForLeader(timeout time.Duration) error {
	return r.store.WaitForLeader(timeout)
}

