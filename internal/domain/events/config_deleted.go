package events

import (
	"encoding/json"
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// ConfigDeleted event is published when a config is deleted
type ConfigDeleted struct {
	EventID         string               `json:"event_id"`
	EventType       string               `json:"event_type"`
	OccurredAt      time.Time            `json:"occurred_at"`
	ProjectID       string               `json:"project_id"`
	ConfigKey       string               `json:"config_key"`
	LastVersion     valueobjects.Version `json:"last_version"`
	DeletedByUserID string               `json:"deleted_by_user_id"`
}

// NewConfigDeleted creates a new ConfigDeleted event
func NewConfigDeleted(
	eventID, projectID, configKey string,
	lastVersion valueobjects.Version,
	deletedByUserID string,
) *ConfigDeleted {
	return &ConfigDeleted{
		EventID:         eventID,
		EventType:       "config.deleted",
		OccurredAt:      time.Now(),
		ProjectID:       projectID,
		ConfigKey:       configKey,
		LastVersion:     lastVersion,
		DeletedByUserID: deletedByUserID,
	}
}

// GetEventID returns the event ID
func (e *ConfigDeleted) GetEventID() string {
	return e.EventID
}

// GetEventType returns the event type
func (e *ConfigDeleted) GetEventType() string {
	return e.EventType
}

// GetOccurredAt returns when the event occurred
func (e *ConfigDeleted) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// ToJSON converts the event to JSON
func (e *ConfigDeleted) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

