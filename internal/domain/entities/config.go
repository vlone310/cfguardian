package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// Config represents the current authoritative configuration state
// This entity requires Raft consensus and uses optimistic locking
type Config struct {
	projectID       string
	key             string
	schemaID        string
	version         valueobjects.Version
	content         json.RawMessage
	updatedByUserID string
	createdAt       time.Time
	updatedAt       time.Time
}

// NewConfig creates a new Config entity with initial version
func NewConfig(projectID, key, schemaID string, content json.RawMessage, updatedByUserID string) *Config {
	now := time.Now()
	return &Config{
		projectID:       projectID,
		key:             key,
		schemaID:        schemaID,
		version:         valueobjects.InitialVersion(),
		content:         content,
		updatedByUserID: updatedByUserID,
		createdAt:       now,
		updatedAt:       now,
	}
}

// ReconstructConfig reconstructs a Config from persistence layer
func ReconstructConfig(
	projectID, key, schemaID string,
	version valueobjects.Version,
	content json.RawMessage,
	updatedByUserID string,
	createdAt, updatedAt time.Time,
) *Config {
	return &Config{
		projectID:       projectID,
		key:             key,
		schemaID:        schemaID,
		version:         version,
		content:         content,
		updatedByUserID: updatedByUserID,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// ProjectID returns the project ID
func (c *Config) ProjectID() string {
	return c.projectID
}

// Key returns the configuration key
func (c *Config) Key() string {
	return c.key
}

// SchemaID returns the schema ID
func (c *Config) SchemaID() string {
	return c.schemaID
}

// Version returns the current version
func (c *Config) Version() valueobjects.Version {
	return c.version
}

// Content returns the configuration content
func (c *Config) Content() json.RawMessage {
	return c.content
}

// UpdatedByUserID returns the ID of the user who last updated this config
func (c *Config) UpdatedByUserID() string {
	return c.updatedByUserID
}

// CreatedAt returns the creation timestamp
func (c *Config) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt returns the last update timestamp
func (c *Config) UpdatedAt() time.Time {
	return c.updatedAt
}

// UpdateContent updates the configuration content with optimistic locking
// Returns error if the expected version doesn't match current version
func (c *Config) UpdateContent(expectedVersion valueobjects.Version, newContent json.RawMessage, updatedByUserID string) error {
	if !c.version.Equals(expectedVersion) {
		return fmt.Errorf(
			"version mismatch: expected %s, current %s (concurrent modification detected)",
			expectedVersion.String(),
			c.version.String(),
		)
	}
	
	c.content = newContent
	c.version = c.version.Next()
	c.updatedByUserID = updatedByUserID
	c.updatedAt = time.Now()
	
	return nil
}

// ChangeSchema updates the schema ID
func (c *Config) ChangeSchema(schemaID, updatedByUserID string) {
	c.schemaID = schemaID
	c.updatedByUserID = updatedByUserID
	c.updatedAt = time.Now()
}

// IsVersion checks if the config is at a specific version
func (c *Config) IsVersion(v valueobjects.Version) bool {
	return c.version.Equals(v)
}

// IsNewerThan checks if this config is newer than a specific version
func (c *Config) IsNewerThan(v valueobjects.Version) bool {
	return c.version.IsGreaterThan(v)
}

// IsInitialVersion checks if this is the initial version
func (c *Config) IsInitialVersion() bool {
	return c.version.IsInitial()
}

// GetContentAsString returns the content as a formatted JSON string
func (c *Config) GetContentAsString() (string, error) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, c.content, "", "  "); err != nil {
		return "", fmt.Errorf("failed to format content: %w", err)
	}
	return pretty.String(), nil
}

// UnmarshalContent unmarshals the content into the provided struct
func (c *Config) UnmarshalContent(v interface{}) error {
	if err := json.Unmarshal(c.content, v); err != nil {
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return nil
}

// Equals checks if two configs are the same entity (by project ID and key)
func (c *Config) Equals(other *Config) bool {
	if other == nil {
		return false
	}
	return c.projectID == other.projectID && c.key == other.key
}

// IsSameVersion checks if two configs are at the same version
func (c *Config) IsSameVersion(other *Config) bool {
	if other == nil {
		return false
	}
	return c.Equals(other) && c.version.Equals(other.version)
}

