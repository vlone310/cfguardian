package entities

import (
	"encoding/json"
	"fmt"
	"time"
)

// ConfigSchema represents a reusable JSON Schema definition
type ConfigSchema struct {
	id              string
	name            string
	schemaContent   string // JSON Schema as string
	createdByUserID string
	createdAt       time.Time
	updatedAt       time.Time
}

// NewConfigSchema creates a new ConfigSchema entity
func NewConfigSchema(id, name, schemaContent, createdByUserID string) (*ConfigSchema, error) {
	// Validate that schemaContent is valid JSON
	if !isValidJSON(schemaContent) {
		return nil, fmt.Errorf("schema content is not valid JSON")
	}
	
	now := time.Now()
	return &ConfigSchema{
		id:              id,
		name:            name,
		schemaContent:   schemaContent,
		createdByUserID: createdByUserID,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// ReconstructConfigSchema reconstructs a ConfigSchema from persistence layer
func ReconstructConfigSchema(
	id, name, schemaContent, createdByUserID string,
	createdAt, updatedAt time.Time,
) *ConfigSchema {
	return &ConfigSchema{
		id:              id,
		name:            name,
		schemaContent:   schemaContent,
		createdByUserID: createdByUserID,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// ID returns the schema ID
func (cs *ConfigSchema) ID() string {
	return cs.id
}

// Name returns the schema name
func (cs *ConfigSchema) Name() string {
	return cs.name
}

// SchemaContent returns the JSON Schema content
func (cs *ConfigSchema) SchemaContent() string {
	return cs.schemaContent
}

// CreatedByUserID returns the ID of the user who created this schema
func (cs *ConfigSchema) CreatedByUserID() string {
	return cs.createdByUserID
}

// CreatedAt returns the creation timestamp
func (cs *ConfigSchema) CreatedAt() time.Time {
	return cs.createdAt
}

// UpdatedAt returns the last update timestamp
func (cs *ConfigSchema) UpdatedAt() time.Time {
	return cs.updatedAt
}

// UpdateName updates the schema name
func (cs *ConfigSchema) UpdateName(name string) {
	cs.name = name
	cs.updatedAt = time.Now()
}

// UpdateSchemaContent updates the schema content
func (cs *ConfigSchema) UpdateSchemaContent(content string) error {
	// Validate that the new content is valid JSON
	if !isValidJSON(content) {
		return fmt.Errorf("schema content is not valid JSON")
	}
	
	cs.schemaContent = content
	cs.updatedAt = time.Now()
	return nil
}

// GetSchemaAsMap returns the schema content as a map for validation
func (cs *ConfigSchema) GetSchemaAsMap() (map[string]interface{}, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(cs.schemaContent), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema content: %w", err)
	}
	return schema, nil
}

// IsCreatedBy checks if the schema was created by a specific user
func (cs *ConfigSchema) IsCreatedBy(userID string) bool {
	return cs.createdByUserID == userID
}

// Equals checks if two schemas are the same entity (by ID)
func (cs *ConfigSchema) Equals(other *ConfigSchema) bool {
	if other == nil {
		return false
	}
	return cs.id == other.id
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

