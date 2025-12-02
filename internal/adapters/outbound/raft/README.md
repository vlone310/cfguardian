# Raft Consensus for GoConfig Guardian

This package implements **Raft consensus** using `hashicorp/raft` to provide **strong consistency (CP)** for configuration data.

## Why Raft?

GoConfig Guardian requires **strong consistency** for configuration changes:

- **No Split Brain**: All nodes agree on the current configuration state
- **Optimistic Locking**: Version conflicts are resolved through consensus
- **Durability**: Configurations are replicated across multiple nodes
- **Leader Election**: Automatic failover if the leader fails

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Node 1    │────▶│   Node 2    │────▶│   Node 3    │
│  (Leader)   │     │  (Follower) │     │  (Follower) │
└─────────────┘     └─────────────┘     └─────────────┘
       │                    │                    │
       ▼                    ▼                    ▼
   [Write to             [Replicate]         [Replicate]
    Raft Log]            [Raft Log]          [Raft Log]
       │                    │                    │
       ▼                    ▼                    ▼
    [Apply to            [Apply to]           [Apply to]
     FSM]                 [FSM]                 [FSM]
```

## Components

### 1. FSM (Finite State Machine) - `fsm.go`

The FSM handles all state transitions:

**Commands:**
- `CREATE_CONFIG` - Create a new config (version = 1)
- `UPDATE_CONFIG` - Update existing config (version++)
- `DELETE_CONFIG` - Delete a config

**State:**
- In-memory map of all configs: `map[string]*ConfigState`
- Key format: `"projectID:configKey"`

**Optimistic Locking:**
```go
// Update fails if version mismatch
cmd := Command{
    Type:            CommandTypeUpdateConfig,
    ExpectedVersion: 5,  // Must match current version!
}
```

### 2. Store - `store.go`

Manages the Raft node lifecycle:

**Responsibilities:**
- Initialize Raft node (transport, log store, snapshots)
- Apply commands through Raft consensus
- Handle leader election
- Manage cluster membership (join/leave)

**Configuration:**
```go
store, err := NewStore(StoreConfig{
    NodeID:           "node1",
    BindAddr:         "127.0.0.1:7000",
    DataDir:          "./raft-data",
    Bootstrap:        true,  // First node bootstraps cluster
    HeartbeatTimeout: 1 * time.Second,
    ElectionTimeout:  1 * time.Second,
    SnapshotInterval: 120 * time.Second,
})
```

### 3. Config Repository - `config_repository.go`

Implements `outbound.ConfigRepository` using Raft:

**Write Operations** (require consensus):
- `Create()` - Goes through Raft leader
- `Update()` - Goes through Raft leader with version check
- `Delete()` - Goes through Raft leader

**Read Operations** (read from local FSM):
- `Get()` - Fast local read (no consensus needed)
- `ListByProject()` - Fast local read

## Consistency Guarantees

### Strong Consistency (CP)

**Writes:**
1. Client sends write to leader
2. Leader appends to Raft log
3. Leader replicates to majority of followers
4. Once majority acknowledges, leader commits
5. Leader applies to FSM
6. Leader responds to client

**Reads:**
- Read from local FSM (linearizable with leader lease)
- **Stale reads possible** on followers (eventual consistency)
- For strict linearizability, read from leader only

### Optimistic Locking

```go
// Concurrent update scenario:
// User A: expectedVersion = 5
// User B: expectedVersion = 5

// User A's update succeeds (version becomes 6)
// User B's update FAILS with version mismatch error
```

## Usage Example

### Initialize Raft Store

```go
// Node 1 (Bootstrap)
store1, err := raft.NewStore(raft.StoreConfig{
    NodeID:    "node1",
    BindAddr:  "192.168.1.1:7000",
    DataDir:   "/var/lib/cfguardian/raft",
    Bootstrap: true,
})

// Wait for leader election
store1.WaitForLeader(10 * time.Second)
```

### Add More Nodes

```go
// Node 2 (Join existing cluster)
store2, err := raft.NewStore(raft.StoreConfig{
    NodeID:    "node2",
    BindAddr:  "192.168.1.2:7000",
    DataDir:   "/var/lib/cfguardian/raft",
    Bootstrap: false,
})

// On leader (node1), add node2 to cluster
store1.Join("node2", "192.168.1.2:7000")
```

### Create Config Repository

```go
// Create Raft-backed config repository
configRepo := raft.NewConfigRepository(store1)

// Check if this node is the leader
if !configRepo.IsLeader() {
    // Redirect to leader
    leaderAddr := configRepo.GetLeaderAddress()
    // ... redirect client to leader
}

// Create config (only works on leader)
config, err := configRepo.Create(ctx, outbound.CreateConfigParams{
    ProjectID:       "proj-1",
    Key:             "app-config",
    SchemaID:        "schema-1",
    Content:         []byte(`{"setting": "value"}`),
    UpdatedByUserID: "user-1",
})
```

### Update Config with Optimistic Locking

```go
// Get current config
current, _ := configRepo.Get(ctx, "proj-1", "app-config")

// Update with version check
updated, err := configRepo.Update(ctx, outbound.UpdateConfigParams{
    ProjectID:       "proj-1",
    Key:             "app-config",
    ExpectedVersion: current.Version,  // ← Optimistic lock!
    Content:         newContent,
    UpdatedByUserID: "user-1",
})

if err != nil {
    // Version mismatch = concurrent modification detected!
    // Client should retry with latest version
}
```

## Deployment Patterns

### Single Node (Development)

```yaml
RAFT_NODE_ID: node1
RAFT_BIND_ADDR: 127.0.0.1:7000
RAFT_BOOTSTRAP: true
```

### 3-Node Cluster (Production)

**Node 1** (Bootstrap):
```yaml
RAFT_NODE_ID: node1
RAFT_BIND_ADDR: 10.0.1.1:7000
RAFT_BOOTSTRAP: true
```

**Node 2** (Join):
```yaml
RAFT_NODE_ID: node2
RAFT_BIND_ADDR: 10.0.1.2:7000
RAFT_JOIN_ADDRESSES: 10.0.1.1:7000
```

**Node 3** (Join):
```yaml
RAFT_NODE_ID: node3
RAFT_BIND_ADDR: 10.0.1.3:7000
RAFT_JOIN_ADDRESSES: 10.0.1.1:7000
```

## Monitoring

### Check Raft Status

```go
// Get Raft stats
stats := store.Stats()

// Important metrics:
// - state: "Leader" | "Follower" | "Candidate"
// - last_log_index: Latest log entry
// - last_applied: Last entry applied to FSM
// - commit_index: Last committed entry
// - num_peers: Number of cluster nodes
```

### Health Checks

```go
// Check if leader exists
if store.GetLeader() == "" {
    // No leader - cluster unavailable
}

// Check if this node is leader
if store.IsLeader() {
    // This node can handle writes
}
```

## Snapshots

Raft automatically creates snapshots to compact the log:

**Configuration:**
- `SnapshotInterval`: How often to check for snapshots (default: 120s)
- `SnapshotThreshold`: Create snapshot after N log entries (default: 8192)
- `TrailingLogs`: Keep N logs after snapshot (default: 10240)

**Snapshot Location:**
```
./raft-data/
├── raft-log.db          # Raft log (BoltDB)
├── raft-stable.db       # Stable store (BoltDB)
└── snapshots/
    ├── meta.json
    └── state.bin        # FSM snapshot (JSON)
```

## Failure Scenarios

### Leader Failure

1. Followers detect missing heartbeats
2. Election timeout triggers
3. Follower becomes candidate
4. Candidate requests votes
5. Majority votes → new leader elected
6. **Automatic failover** (no data loss)

### Network Partition

- **Majority partition**: Continues operating (CP system)
- **Minority partition**: Cannot commit writes (unavailable)
- Once partition heals, minority catches up via log replication

### Split Brain Prevention

- Requires **majority** (quorum) for all operations
- 3-node cluster: needs 2 nodes
- 5-node cluster: needs 3 nodes
- **Impossible** to have two leaders simultaneously

## Performance Considerations

### Writes

- **Latency**: ~2-10ms (depends on network + replication)
- **Throughput**: Limited by leader (single leader writes)
- **Optimization**: Batch multiple configs in single Raft entry

### Reads

- **Local FSM reads**: <1ms (no network)
- **Linearizable reads**: Add leader lease check (~1ms)
- **Optimization**: Read from followers for stale-read tolerance

### Scaling

- **Vertical**: Faster disks for log writes
- **Horizontal**: Use PostgreSQL for read-heavy queries
- **Hybrid**: Raft for writes, PostgreSQL for complex queries

## Troubleshooting

### "not the leader" Error

**Solution**: Redirect client to current leader

```go
if err.Error() == "not the leader" {
    leaderAddr := configRepo.GetLeaderAddress()
    // Return 307 Temporary Redirect to client
}
```

### Cluster Won't Elect Leader

**Check:**
1. Network connectivity between nodes
2. Correct bind addresses configured
3. Firewall allows Raft port (default: 7000)
4. Logs for election timeout/vote issues

### Snapshot Restore Fails

**Check:**
1. Snapshot file integrity
2. Disk space available
3. FSM version compatibility
4. Logs for specific error

## Future Enhancements

- [ ] Support for read-only followers (reduce leader load)
- [ ] Multi-Raft for horizontal scaling (shard by project)
- [ ] Batch operations (multiple configs in one Raft entry)
- [ ] Compression for snapshots
- [ ] Metrics export (Prometheus)

