package events

import (
	"time"
)

// DomainEvent represents a domain event interface
// All domain events must implement this interface
type DomainEvent interface {
	// GetEventID returns the unique event identifier
	GetEventID() string
	
	// GetEventType returns the event type (e.g., "config.created")
	GetEventType() string
	
	// GetOccurredAt returns when the event occurred
	GetOccurredAt() time.Time
	
	// ToJSON converts the event to JSON for serialization
	ToJSON() ([]byte, error)
}

// EventType constants for all domain events
const (
	EventTypeConfigCreated    = "config.created"
	EventTypeConfigUpdated    = "config.updated"
	EventTypeConfigDeleted    = "config.deleted"
	EventTypeConfigRolledBack = "config.rolledback"
	EventTypeUserCreated      = "user.created"
	EventTypeUserDeleted      = "user.deleted"
	EventTypeProjectCreated   = "project.created"
	EventTypeProjectDeleted   = "project.deleted"
	EventTypeRoleAssigned     = "role.assigned"
	EventTypeRoleRevoked      = "role.revoked"
	EventTypeSchemaCreated    = "schema.created"
	EventTypeSchemaUpdated    = "schema.updated"
	EventTypeSchemaDeleted    = "schema.deleted"
)

// Verify that our event types implement the DomainEvent interface
var (
	_ DomainEvent = (*ConfigCreated)(nil)
	_ DomainEvent = (*ConfigUpdated)(nil)
	_ DomainEvent = (*ConfigDeleted)(nil)
	_ DomainEvent = (*ConfigRolledBack)(nil)
)

