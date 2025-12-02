package events

import (
	"encoding/json"
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// ConfigRolledBack event is published when a config is rolled back to a previous version
type ConfigRolledBack struct {
	EventID           string               `json:"event_id"`
	EventType         string               `json:"event_type"`
	OccurredAt        time.Time            `json:"occurred_at"`
	ProjectID         string               `json:"project_id"`
	ConfigKey         string               `json:"config_key"`
	FromVersion       valueobjects.Version `json:"from_version"`
	ToVersion         valueobjects.Version `json:"to_version"`
	Content           json.RawMessage      `json:"content"`
	RolledBackByUserID string              `json:"rolled_back_by_user_id"`
}

// NewConfigRolledBack creates a new ConfigRolledBack event
func NewConfigRolledBack(
	eventID, projectID, configKey string,
	fromVersion, toVersion valueobjects.Version,
	content json.RawMessage,
	rolledBackByUserID string,
) *ConfigRolledBack {
	return &ConfigRolledBack{
		EventID:           eventID,
		EventType:         "config.rolledback",
		OccurredAt:        time.Now(),
		ProjectID:         projectID,
		ConfigKey:         configKey,
		FromVersion:       fromVersion,
		ToVersion:         toVersion,
		Content:           content,
		RolledBackByUserID: rolledBackByUserID,
	}
}

// GetEventID returns the event ID
func (e *ConfigRolledBack) GetEventID() string {
	return e.EventID
}

// GetEventType returns the event type
func (e *ConfigRolledBack) GetEventType() string {
	return e.EventType
}

// GetOccurredAt returns when the event occurred
func (e *ConfigRolledBack) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// ToJSON converts the event to JSON
func (e *ConfigRolledBack) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

