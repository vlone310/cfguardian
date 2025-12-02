package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// ConfigRevision represents an immutable historical record of a configuration
// This is part of the audit log and should never be modified after creation
type ConfigRevision struct {
	id              string
	projectID       string
	configKey       string
	version         valueobjects.Version
	content         json.RawMessage
	createdByUserID string
	createdAt       time.Time
}

// NewConfigRevision creates a new ConfigRevision entity
func NewConfigRevision(
	id, projectID, configKey string,
	version valueobjects.Version,
	content json.RawMessage,
	createdByUserID string,
) *ConfigRevision {
	return &ConfigRevision{
		id:              id,
		projectID:       projectID,
		configKey:       configKey,
		version:         version,
		content:         content,
		createdByUserID: createdByUserID,
		createdAt:       time.Now(),
	}
}

// ReconstructConfigRevision reconstructs a ConfigRevision from persistence layer
func ReconstructConfigRevision(
	id, projectID, configKey string,
	version valueobjects.Version,
	content json.RawMessage,
	createdByUserID string,
	createdAt time.Time,
) *ConfigRevision {
	return &ConfigRevision{
		id:              id,
		projectID:       projectID,
		configKey:       configKey,
		version:         version,
		content:         content,
		createdByUserID: createdByUserID,
		createdAt:       createdAt,
	}
}

// ID returns the revision ID
func (cr *ConfigRevision) ID() string {
	return cr.id
}

// ProjectID returns the project ID
func (cr *ConfigRevision) ProjectID() string {
	return cr.projectID
}

// ConfigKey returns the configuration key
func (cr *ConfigRevision) ConfigKey() string {
	return cr.configKey
}

// Version returns the version number
func (cr *ConfigRevision) Version() valueobjects.Version {
	return cr.version
}

// Content returns the configuration content snapshot
func (cr *ConfigRevision) Content() json.RawMessage {
	return cr.content
}

// CreatedByUserID returns the ID of the user who created this revision
func (cr *ConfigRevision) CreatedByUserID() string {
	return cr.createdByUserID
}

// CreatedAt returns the creation timestamp
func (cr *ConfigRevision) CreatedAt() time.Time {
	return cr.createdAt
}

// GetContentAsString returns the content as a formatted JSON string
func (cr *ConfigRevision) GetContentAsString() (string, error) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, cr.content, "", "  "); err != nil {
		return "", fmt.Errorf("failed to format content: %w", err)
	}
	return pretty.String(), nil
}

// UnmarshalContent unmarshals the content into the provided struct
func (cr *ConfigRevision) UnmarshalContent(v interface{}) error {
	if err := json.Unmarshal(cr.content, v); err != nil {
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return nil
}

// IsVersion checks if this revision is a specific version
func (cr *ConfigRevision) IsVersion(v valueobjects.Version) bool {
	return cr.version.Equals(v)
}

// IsNewerThan checks if this revision is newer than a specific version
func (cr *ConfigRevision) IsNewerThan(v valueobjects.Version) bool {
	return cr.version.IsGreaterThan(v)
}

// IsOlderThan checks if this revision is older than a specific version
func (cr *ConfigRevision) IsOlderThan(v valueobjects.Version) bool {
	return cr.version.IsLessThan(v)
}

// BelongsToConfig checks if this revision belongs to a specific config
func (cr *ConfigRevision) BelongsToConfig(projectID, configKey string) bool {
	return cr.projectID == projectID && cr.configKey == configKey
}

// IsCreatedBy checks if the revision was created by a specific user
func (cr *ConfigRevision) IsCreatedBy(userID string) bool {
	return cr.createdByUserID == userID
}

// Equals checks if two revisions are the same entity (by ID)
func (cr *ConfigRevision) Equals(other *ConfigRevision) bool {
	if other == nil {
		return false
	}
	return cr.id == other.id
}

// IsSameVersion checks if two revisions are at the same version for the same config
func (cr *ConfigRevision) IsSameVersion(other *ConfigRevision) bool {
	if other == nil {
		return false
	}
	return cr.projectID == other.projectID &&
		cr.configKey == other.configKey &&
		cr.version.Equals(other.version)
}

