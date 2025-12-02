package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

// Store manages the Raft consensus and provides config operations
type Store struct {
	raft *raft.Raft
	fsm  *FSM
	
	// Configuration
	nodeID      string
	bindAddr    string
	dataDir     string
	localID     raft.ServerID
	localAddr   raft.ServerAddress
}

// StoreConfig holds Raft store configuration
type StoreConfig struct {
	NodeID               string
	BindAddr             string
	DataDir              string
	Bootstrap            bool
	HeartbeatTimeout     time.Duration
	ElectionTimeout      time.Duration
	SnapshotInterval     time.Duration
	SnapshotThreshold    uint64
	TrailingLogs         uint64
}

// NewStore creates a new Raft store
func NewStore(cfg StoreConfig) (*Store, error) {
	// Validate configuration
	if cfg.NodeID == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if cfg.BindAddr == "" {
		return nil, fmt.Errorf("bind address is required")
	}
	if cfg.DataDir == "" {
		return nil, fmt.Errorf("data directory is required")
	}
	
	// Set defaults
	if cfg.HeartbeatTimeout == 0 {
		cfg.HeartbeatTimeout = 1 * time.Second
	}
	if cfg.ElectionTimeout == 0 {
		cfg.ElectionTimeout = 1 * time.Second
	}
	if cfg.SnapshotInterval == 0 {
		cfg.SnapshotInterval = 120 * time.Second
	}
	if cfg.SnapshotThreshold == 0 {
		cfg.SnapshotThreshold = 8192
	}
	if cfg.TrailingLogs == 0 {
		cfg.TrailingLogs = 10240
	}
	
	store := &Store{
		nodeID:    cfg.NodeID,
		bindAddr:  cfg.BindAddr,
		dataDir:   cfg.DataDir,
		localID:   raft.ServerID(cfg.NodeID),
		localAddr: raft.ServerAddress(cfg.BindAddr),
	}
	
	// Create FSM
	store.fsm = NewFSM()
	
	// Initialize Raft
	if err := store.initRaft(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize raft: %w", err)
	}
	
	return store, nil
}

// initRaft initializes the Raft node
func (s *Store) initRaft(cfg StoreConfig) error {
	// Create data directory
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Setup Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = s.localID
	config.HeartbeatTimeout = cfg.HeartbeatTimeout
	config.ElectionTimeout = cfg.ElectionTimeout
	config.SnapshotInterval = cfg.SnapshotInterval
	config.SnapshotThreshold = cfg.SnapshotThreshold
	config.TrailingLogs = cfg.TrailingLogs
	
	// Setup transport
	addr, err := net.ResolveTCPAddr("tcp", s.bindAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve bind address: %w", err)
	}
	
	transport, err := raft.NewTCPTransport(s.bindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}
	
	// Setup log store (BoltDB)
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.dataDir, "raft-log.db"))
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}
	
	// Setup stable store (BoltDB)
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(s.dataDir, "raft-stable.db"))
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}
	
	// Setup snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(s.dataDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %w", err)
	}
	
	// Create the Raft node
	ra, err := raft.NewRaft(config, s.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return fmt.Errorf("failed to create raft: %w", err)
	}
	
	s.raft = ra
	
	// Bootstrap cluster if needed
	if cfg.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      s.localID,
					Address: s.localAddr,
				},
			},
		}
		s.raft.BootstrapCluster(configuration)
	}
	
	return nil
}

// CreateConfig creates a new config through Raft consensus
func (s *Store) CreateConfig(ctx context.Context, projectID, key, schemaID string, content json.RawMessage, userID string) (*ConfigState, error) {
	if !s.IsLeader() {
		return nil, fmt.Errorf("not the leader")
	}
	
	cmd := Command{
		Type:            CommandTypeCreateConfig,
		ProjectID:       projectID,
		Key:             key,
		SchemaID:        schemaID,
		Content:         content,
		UpdatedByUserID: userID,
	}
	
	return s.applyCommand(ctx, cmd)
}

// UpdateConfig updates an existing config through Raft consensus
func (s *Store) UpdateConfig(ctx context.Context, projectID, key string, expectedVersion int64, content json.RawMessage, userID string) (*ConfigState, error) {
	if !s.IsLeader() {
		return nil, fmt.Errorf("not the leader")
	}
	
	cmd := Command{
		Type:            CommandTypeUpdateConfig,
		ProjectID:       projectID,
		Key:             key,
		Content:         content,
		ExpectedVersion: expectedVersion,
		UpdatedByUserID: userID,
	}
	
	return s.applyCommand(ctx, cmd)
}

// DeleteConfig deletes a config through Raft consensus
func (s *Store) DeleteConfig(ctx context.Context, projectID, key, userID string) error {
	if !s.IsLeader() {
		return fmt.Errorf("not the leader")
	}
	
	cmd := Command{
		Type:            CommandTypeDeleteConfig,
		ProjectID:       projectID,
		Key:             key,
		UpdatedByUserID: userID,
	}
	
	_, err := s.applyCommand(ctx, cmd)
	return err
}

// applyCommand applies a command through Raft consensus
func (s *Store) applyCommand(ctx context.Context, cmd Command) (*ConfigState, error) {
	// Serialize command
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	
	// Apply through Raft with timeout
	timeout := 10 * time.Second
	future := s.raft.Apply(data, timeout)
	
	// Wait for result
	if err := future.Error(); err != nil {
		return nil, fmt.Errorf("raft apply failed: %w", err)
	}
	
	// Get response
	response := future.Response()
	if err, ok := response.(error); ok {
		return nil, err
	}
	
	if config, ok := response.(*ConfigState); ok {
		return config, nil
	}
	
	return nil, nil
}

// GetConfig retrieves a config (read from FSM, no consensus needed)
func (s *Store) GetConfig(projectID, key string) (*ConfigState, error) {
	return s.fsm.GetConfig(projectID, key)
}

// ListConfigs lists configs for a project (read from FSM)
func (s *Store) ListConfigs(projectID string) []*ConfigState {
	return s.fsm.ListConfigs(projectID)
}

// IsLeader checks if this node is the Raft leader
func (s *Store) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

// GetLeader returns the current leader address
func (s *Store) GetLeader() string {
	_, leaderID := s.raft.LeaderWithID()
	return string(leaderID)
}

// WaitForLeader waits until a leader is elected
func (s *Store) WaitForLeader(timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	
	for {
		select {
		case <-ticker.C:
			if s.GetLeader() != "" {
				return nil
			}
		case <-timer.C:
			return fmt.Errorf("timeout waiting for leader")
		}
	}
}

// Join adds a new node to the Raft cluster
func (s *Store) Join(nodeID, addr string) error {
	if !s.IsLeader() {
		return fmt.Errorf("not the leader")
	}
	
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return fmt.Errorf("failed to get raft configuration: %w", err)
	}
	
	// Check if node already exists
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) {
			// Node already exists, remove it first
			removeFuture := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := removeFuture.Error(); err != nil {
				return fmt.Errorf("failed to remove existing server: %w", err)
			}
		}
	}
	
	// Add the node
	addFuture := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if err := addFuture.Error(); err != nil {
		return fmt.Errorf("failed to add voter: %w", err)
	}
	
	return nil
}

// Leave removes a node from the Raft cluster
func (s *Store) Leave(nodeID string) error {
	if !s.IsLeader() {
		return fmt.Errorf("not the leader")
	}
	
	removeFuture := s.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
	if err := removeFuture.Error(); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}
	
	return nil
}

// Shutdown gracefully shuts down the Raft node
func (s *Store) Shutdown() error {
	future := s.raft.Shutdown()
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to shutdown raft: %w", err)
	}
	return nil
}

// Stats returns Raft stats
func (s *Store) Stats() map[string]string {
	return s.raft.Stats()
}

