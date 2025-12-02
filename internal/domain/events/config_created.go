package events

import (
	"encoding/json"
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// ConfigCreated event is published when a new config is created
type ConfigCreated struct {
	EventID         string                  `json:"event_id"`
	EventType       string                  `json:"event_type"`
	OccurredAt      time.Time               `json:"occurred_at"`
	ProjectID       string                  `json:"project_id"`
	ConfigKey       string                  `json:"config_key"`
	SchemaID        string                  `json:"schema_id"`
	Version         valueobjects.Version    `json:"version"`
	Content         json.RawMessage         `json:"content"`
	CreatedByUserID string                  `json:"created_by_user_id"`
}

// NewConfigCreated creates a new ConfigCreated event
func NewConfigCreated(
	eventID, projectID, configKey, schemaID string,
	version valueobjects.Version,
	content json.RawMessage,
	createdByUserID string,
) *ConfigCreated {
	return &ConfigCreated{
		EventID:         eventID,
		EventType:       "config.created",
		OccurredAt:      time.Now(),
		ProjectID:       projectID,
		ConfigKey:       configKey,
		SchemaID:        schemaID,
		Version:         version,
		Content:         content,
		CreatedByUserID: createdByUserID,
	}
}

// GetEventID returns the event ID
func (e *ConfigCreated) GetEventID() string {
	return e.EventID
}

// GetEventType returns the event type
func (e *ConfigCreated) GetEventType() string {
	return e.EventType
}

// GetOccurredAt returns when the event occurred
func (e *ConfigCreated) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// ToJSON converts the event to JSON
func (e *ConfigCreated) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

