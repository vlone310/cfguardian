package events

import (
	"encoding/json"
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// ConfigUpdated event is published when a config is updated
type ConfigUpdated struct {
	EventID         string                  `json:"event_id"`
	EventType       string                  `json:"event_type"`
	OccurredAt      time.Time               `json:"occurred_at"`
	ProjectID       string                  `json:"project_id"`
	ConfigKey       string                  `json:"config_key"`
	SchemaID        string                  `json:"schema_id"`
	PreviousVersion valueobjects.Version    `json:"previous_version"`
	NewVersion      valueobjects.Version    `json:"new_version"`
	Content         json.RawMessage         `json:"content"`
	UpdatedByUserID string                  `json:"updated_by_user_id"`
}

// NewConfigUpdated creates a new ConfigUpdated event
func NewConfigUpdated(
	eventID, projectID, configKey, schemaID string,
	previousVersion, newVersion valueobjects.Version,
	content json.RawMessage,
	updatedByUserID string,
) *ConfigUpdated {
	return &ConfigUpdated{
		EventID:         eventID,
		EventType:       "config.updated",
		OccurredAt:      time.Now(),
		ProjectID:       projectID,
		ConfigKey:       configKey,
		SchemaID:        schemaID,
		PreviousVersion: previousVersion,
		NewVersion:      newVersion,
		Content:         content,
		UpdatedByUserID: updatedByUserID,
	}
}

// GetEventID returns the event ID
func (e *ConfigUpdated) GetEventID() string {
	return e.EventID
}

// GetEventType returns the event type
func (e *ConfigUpdated) GetEventType() string {
	return e.EventType
}

// GetOccurredAt returns when the event occurred
func (e *ConfigUpdated) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// ToJSON converts the event to JSON
func (e *ConfigUpdated) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

