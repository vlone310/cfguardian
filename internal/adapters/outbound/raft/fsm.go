package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// CommandType represents the type of Raft command
type CommandType string

const (
	CommandTypeCreateConfig CommandType = "CREATE_CONFIG"
	CommandTypeUpdateConfig CommandType = "UPDATE_CONFIG"
	CommandTypeDeleteConfig CommandType = "DELETE_CONFIG"
)

// Command represents a Raft log command
type Command struct {
	Type            CommandType     `json:"type"`
	ProjectID       string          `json:"project_id"`
	Key             string          `json:"key"`
	SchemaID        string          `json:"schema_id,omitempty"`
	Content         json.RawMessage `json:"content,omitempty"`
	ExpectedVersion int64           `json:"expected_version,omitempty"`
	UpdatedByUserID string          `json:"updated_by_user_id"`
}

// ConfigState represents the in-memory state of a config
type ConfigState struct {
	ProjectID       string          `json:"project_id"`
	Key             string          `json:"key"`
	SchemaID        string          `json:"schema_id"`
	Version         int64           `json:"version"`
	Content         json.RawMessage `json:"content"`
	UpdatedByUserID string          `json:"updated_by_user_id"`
}

// FSM implements the Raft Finite State Machine
// This is where all state changes happen
type FSM struct {
	mu      sync.RWMutex
	configs map[string]*ConfigState // key: "projectID:configKey"
}

// NewFSM creates a new FSM
func NewFSM() *FSM {
	return &FSM{
		configs: make(map[string]*ConfigState),
	}
}

// Apply applies a Raft log entry to the FSM
// This is called by Raft when a log entry is committed
func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal command: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Type {
	case CommandTypeCreateConfig:
		return f.applyCreateConfig(cmd)
	case CommandTypeUpdateConfig:
		return f.applyUpdateConfig(cmd)
	case CommandTypeDeleteConfig:
		return f.applyDeleteConfig(cmd)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

// applyCreateConfig creates a new config in the FSM
func (f *FSM) applyCreateConfig(cmd Command) interface{} {
	key := makeKey(cmd.ProjectID, cmd.Key)
	
	// Check if config already exists
	if _, exists := f.configs[key]; exists {
		return fmt.Errorf("config already exists: %s", key)
	}
	
	// Create new config with version 1
	config := &ConfigState{
		ProjectID:       cmd.ProjectID,
		Key:             cmd.Key,
		SchemaID:        cmd.SchemaID,
		Version:         1,
		Content:         cmd.Content,
		UpdatedByUserID: cmd.UpdatedByUserID,
	}
	
	f.configs[key] = config
	return config
}

// applyUpdateConfig updates an existing config with optimistic locking
func (f *FSM) applyUpdateConfig(cmd Command) interface{} {
	key := makeKey(cmd.ProjectID, cmd.Key)
	
	// Get existing config
	config, exists := f.configs[key]
	if !exists {
		return fmt.Errorf("config not found: %s", key)
	}
	
	// Optimistic locking check
	if config.Version != cmd.ExpectedVersion {
		return fmt.Errorf("version mismatch: expected %d, got %d", cmd.ExpectedVersion, config.Version)
	}
	
	// Update config
	config.Content = cmd.Content
	config.Version++
	config.UpdatedByUserID = cmd.UpdatedByUserID
	
	return config
}

// applyDeleteConfig deletes a config from the FSM
func (f *FSM) applyDeleteConfig(cmd Command) interface{} {
	key := makeKey(cmd.ProjectID, cmd.Key)
	
	// Check if config exists
	if _, exists := f.configs[key]; !exists {
		return fmt.Errorf("config not found: %s", key)
	}
	
	// Delete config
	delete(f.configs, key)
	return nil
}

// Snapshot returns a snapshot of the FSM state
// This is called by Raft to create a snapshot
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	// Clone the configs map
	clone := make(map[string]*ConfigState, len(f.configs))
	for k, v := range f.configs {
		// Deep copy
		clone[k] = &ConfigState{
			ProjectID:       v.ProjectID,
			Key:             v.Key,
			SchemaID:        v.SchemaID,
			Version:         v.Version,
			Content:         append(json.RawMessage(nil), v.Content...),
			UpdatedByUserID: v.UpdatedByUserID,
		}
	}
	
	return &FSMSnapshot{configs: clone}, nil
}

// Restore restores the FSM state from a snapshot
// This is called by Raft when restoring from a snapshot
func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	
	var configs map[string]*ConfigState
	if err := json.NewDecoder(rc).Decode(&configs); err != nil {
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}
	
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.configs = configs
	return nil
}

// GetConfig retrieves a config from the FSM (read-only)
func (f *FSM) GetConfig(projectID, key string) (*ConfigState, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	configKey := makeKey(projectID, key)
	config, exists := f.configs[configKey]
	if !exists {
		return nil, fmt.Errorf("config not found")
	}
	
	return config, nil
}

// ListConfigs lists all configs for a project (read-only)
func (f *FSM) ListConfigs(projectID string) []*ConfigState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	var configs []*ConfigState
	for _, config := range f.configs {
		if config.ProjectID == projectID {
			configs = append(configs, config)
		}
	}
	
	return configs
}

// ConfigExists checks if a config exists
func (f *FSM) ConfigExists(projectID, key string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	configKey := makeKey(projectID, key)
	_, exists := f.configs[configKey]
	return exists
}

// GetVersion gets the current version of a config
func (f *FSM) GetVersion(projectID, key string) (valueobjects.Version, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	configKey := makeKey(projectID, key)
	config, exists := f.configs[configKey]
	if !exists {
		return valueobjects.Version{}, fmt.Errorf("config not found")
	}
	
	return valueobjects.MustNewVersion(config.Version), nil
}

// makeKey creates a composite key from projectID and configKey
func makeKey(projectID, configKey string) string {
	return projectID + ":" + configKey
}

// FSMSnapshot implements raft.FSMSnapshot
type FSMSnapshot struct {
	configs map[string]*ConfigState
}

// Persist writes the snapshot to the given sink
func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode the snapshot
		if err := json.NewEncoder(sink).Encode(s.configs); err != nil {
			return fmt.Errorf("failed to encode snapshot: %w", err)
		}
		return nil
	}()
	
	if err != nil {
		sink.Cancel()
		return err
	}
	
	return sink.Close()
}

// Release is called when the snapshot is no longer needed
func (s *FSMSnapshot) Release() {
	// Nothing to release in our case
}

