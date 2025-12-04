package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator_Validate(t *testing.T) {
	validator := NewSchemaValidator()

	tests := []struct {
		name          string
		schema        string
		content       json.RawMessage
		expectValid   bool
		expectErrors  int
		errorContains string
	}{
		{
			name: "valid simple schema and content",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer"}
				},
				"required": ["name"]
			}`,
			content: json.RawMessage(`{
				"name": "John Doe",
				"age": 30
			}`),
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "missing required field",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"email": {"type": "string"}
				},
				"required": ["name", "email"]
			}`,
			content: json.RawMessage(`{
				"name": "John Doe"
			}`),
			expectValid:   false,
			expectErrors:  1,
			errorContains: "email",
		},
		{
			name: "wrong type for field",
			schema: `{
				"type": "object",
				"properties": {
					"age": {"type": "integer"}
				}
			}`,
			content: json.RawMessage(`{
				"age": "thirty"
			}`),
			expectValid:   false,
			expectErrors:  1,
			errorContains: "Invalid type",
		},
		{
			name: "additional properties not allowed",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				},
				"additionalProperties": false
			}`,
			content: json.RawMessage(`{
				"name": "John",
				"extra": "field"
			}`),
			expectValid:   false,
			expectErrors:  1,
			errorContains: "Additional property",
		},
		{
			name: "array validation",
			schema: `{
				"type": "array",
				"items": {
					"type": "string"
				},
				"minItems": 1
			}`,
			content:      json.RawMessage(`["item1", "item2", "item3"]`),
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "array with wrong item type",
			schema: `{
				"type": "array",
				"items": {
					"type": "number"
				}
			}`,
			content:       json.RawMessage(`[1, 2, "three"]`),
			expectValid:   false,
			expectErrors:  1,
			errorContains: "Invalid type",
		},
		{
			name: "nested object validation",
			schema: `{
				"type": "object",
				"properties": {
					"user": {
						"type": "object",
						"properties": {
							"name": {"type": "string"},
							"address": {
								"type": "object",
								"properties": {
									"city": {"type": "string"},
									"zipcode": {"type": "string"}
								},
								"required": ["city"]
							}
						},
						"required": ["name", "address"]
					}
				}
			}`,
			content: json.RawMessage(`{
				"user": {
					"name": "John",
					"address": {
						"city": "New York",
						"zipcode": "10001"
					}
				}
			}`),
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "enum validation",
			schema: `{
				"type": "object",
				"properties": {
					"status": {
						"type": "string",
						"enum": ["active", "inactive", "pending"]
					}
				}
			}`,
			content: json.RawMessage(`{
				"status": "active"
			}`),
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "enum validation failure",
			schema: `{
				"type": "object",
				"properties": {
					"status": {
						"type": "string",
						"enum": ["active", "inactive", "pending"]
					}
				}
			}`,
			content: json.RawMessage(`{
				"status": "unknown"
			}`),
			expectValid:   false,
			expectErrors:  1,
			errorContains: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := validator.Validate(tt.schema, tt.content)

			// Assert
			require.NoError(t, err, "Validation should not return error")
			assert.Equal(t, tt.expectValid, result.Valid, "Validation result validity mismatch")
			assert.Len(t, result.Errors, tt.expectErrors, "Number of validation errors mismatch")

			if !tt.expectValid && tt.errorContains != "" {
				// Check that at least one error contains the expected string
				found := false
				for _, validationErr := range result.Errors {
					if contains(validationErr.Description, tt.errorContains) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error message to contain: %s", tt.errorContains)
			}
		})
	}
}

func TestSchemaValidator_ValidateOrError(t *testing.T) {
	validator := NewSchemaValidator()

	tests := []struct {
		name        string
		schema      string
		content     json.RawMessage
		expectError bool
	}{
		{
			name: "valid content returns no error",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				}
			}`,
			content:     json.RawMessage(`{"name": "John"}`),
			expectError: false,
		},
		{
			name: "invalid content returns error",
			schema: `{
				"type": "object",
				"properties": {
					"age": {"type": "integer"}
				}
			}`,
			content:     json.RawMessage(`{"age": "not a number"}`),
			expectError: true,
		},
		{
			name:        "invalid schema returns error",
			schema:      `{"type": "invalid_type"}`,
			content:     json.RawMessage(`{}`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.ValidateOrError(tt.schema, tt.content)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchemaValidator_ValidateSchema(t *testing.T) {
	validator := NewSchemaValidator()

	tests := []struct {
		name        string
		schema      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid simple schema",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				}
			}`,
			expectError: false,
		},
		{
			name: "valid complex schema",
			schema: `{
				"type": "object",
				"properties": {
					"user": {
						"type": "object",
						"properties": {
							"id": {"type": "integer"},
							"email": {"type": "string", "format": "email"},
							"roles": {
								"type": "array",
								"items": {"type": "string"}
							}
						},
						"required": ["id", "email"]
					}
				}
			}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			schema:      `{invalid json`,
			expectError: true,
			errorMsg:    "not valid JSON",
		},
		{
			name:        "empty schema",
			schema:      ``,
			expectError: true,
			errorMsg:    "not valid JSON",
		},
		{
			name: "schema with $schema keyword",
			schema: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				}
			}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.ValidateSchema(tt.schema)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchemaValidator_ValidateContent(t *testing.T) {
	validator := NewSchemaValidator()

	tests := []struct {
		name        string
		content     json.RawMessage
		expectError bool
	}{
		{
			name:        "valid object",
			content:     json.RawMessage(`{"name": "John", "age": 30}`),
			expectError: false,
		},
		{
			name:        "valid array",
			content:     json.RawMessage(`[1, 2, 3]`),
			expectError: false,
		},
		{
			name:        "valid string",
			content:     json.RawMessage(`"hello"`),
			expectError: false,
		},
		{
			name:        "valid number",
			content:     json.RawMessage(`42`),
			expectError: false,
		},
		{
			name:        "valid boolean",
			content:     json.RawMessage(`true`),
			expectError: false,
		},
		{
			name:        "valid null",
			content:     json.RawMessage(`null`),
			expectError: false,
		},
		{
			name:        "invalid JSON",
			content:     json.RawMessage(`{invalid`),
			expectError: true,
		},
		{
			name:        "empty content",
			content:     json.RawMessage(``),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.ValidateContent(tt.content)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not valid JSON")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	// Arrange
	ve := ValidationError{
		Field:   "user.email",
		Message: "Invalid format",
		Value:   "not-an-email",
	}

	// Act
	errMsg := ve.Error()

	// Assert
	assert.Contains(t, errMsg, "user.email")
	assert.Contains(t, errMsg, "Invalid format")
}

func TestSchemaValidator_RealWorldScenario(t *testing.T) {
	// This test simulates a real configuration validation scenario
	validator := NewSchemaValidator()

	// Define a schema for a database configuration
	dbConfigSchema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"host": {
				"type": "string",
				"minLength": 1
			},
			"port": {
				"type": "integer",
				"minimum": 1,
				"maximum": 65535
			},
			"database": {
				"type": "string",
				"minLength": 1
			},
			"ssl": {
				"type": "boolean"
			},
			"poolSize": {
				"type": "integer",
				"minimum": 1,
				"maximum": 100,
				"default": 10
			}
		},
		"required": ["host", "port", "database"],
		"additionalProperties": false
	}`

	// Test valid configuration
	t.Run("valid database config", func(t *testing.T) {
		validConfig := json.RawMessage(`{
			"host": "localhost",
			"port": 5432,
			"database": "myapp",
			"ssl": true,
			"poolSize": 20
		}`)

		err := validator.ValidateOrError(dbConfigSchema, validConfig)
		assert.NoError(t, err)
	})

	// Test invalid configuration (missing required field)
	t.Run("missing required field", func(t *testing.T) {
		invalidConfig := json.RawMessage(`{
			"host": "localhost",
			"port": 5432
		}`)

		err := validator.ValidateOrError(dbConfigSchema, invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database")
	})

	// Test invalid configuration (out of range port)
	t.Run("port out of range", func(t *testing.T) {
		invalidConfig := json.RawMessage(`{
			"host": "localhost",
			"port": 99999,
			"database": "myapp"
		}`)

		err := validator.ValidateOrError(dbConfigSchema, invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	// Test invalid configuration (additional property)
	t.Run("additional property not allowed", func(t *testing.T) {
		invalidConfig := json.RawMessage(`{
			"host": "localhost",
			"port": 5432,
			"database": "myapp",
			"timeout": 30
		}`)

		err := validator.ValidateOrError(dbConfigSchema, invalidConfig)
		assert.Error(t, err)
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
