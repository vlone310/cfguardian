package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// ValidationError represents a schema validation error
type ValidationError struct {
	Field       string
	Message     string
	Value       interface{}
	Description string
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ValidationResult holds the result of schema validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// SchemaValidator validates configuration content against JSON Schemas
type SchemaValidator struct{}

// NewSchemaValidator creates a new SchemaValidator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// Validate validates content against a JSON Schema
func (sv *SchemaValidator) Validate(schemaContent string, content json.RawMessage) (*ValidationResult, error) {
	// Parse schema
	schemaLoader := gojsonschema.NewStringLoader(schemaContent)
	
	// Parse content
	contentLoader := gojsonschema.NewBytesLoader(content)
	
	// Perform validation
	result, err := gojsonschema.Validate(schemaLoader, contentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Build validation result
	validationResult := &ValidationResult{
		Valid:  result.Valid(),
		Errors: make([]ValidationError, 0),
	}
	
	// Collect errors if validation failed
	if !result.Valid() {
		for _, err := range result.Errors() {
			validationResult.Errors = append(validationResult.Errors, ValidationError{
				Field:       err.Field(),
				Message:     err.Description(),
				Value:       err.Value(),
				Description: err.String(),
			})
		}
	}
	
	return validationResult, nil
}

// ValidateOrError validates and returns an error if validation fails
func (sv *SchemaValidator) ValidateOrError(schemaContent string, content json.RawMessage) error {
	result, err := sv.Validate(schemaContent, content)
	if err != nil {
		return err
	}
	
	if !result.Valid {
		// Build comprehensive error message
		errMsg := fmt.Sprintf("validation failed with %d error(s):", len(result.Errors))
		for i, ve := range result.Errors {
			errMsg += fmt.Sprintf("\n  %d. %s", i+1, ve.Error())
		}
		return errors.New(errMsg)
	}
	
	return nil
}

// ValidateSchema validates that a schema itself is a valid JSON Schema
func (sv *SchemaValidator) ValidateSchema(schemaContent string) error {
	// Parse as JSON first
	var schema interface{}
	if err := json.Unmarshal([]byte(schemaContent), &schema); err != nil {
		return fmt.Errorf("schema is not valid JSON: %w", err)
	}
	
	// Try to load as JSON Schema
	schemaLoader := gojsonschema.NewStringLoader(schemaContent)
	_, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("invalid JSON Schema: %w", err)
	}
	
	return nil
}

// ValidateContent validates that content is valid JSON
func (sv *SchemaValidator) ValidateContent(content json.RawMessage) error {
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("content is not valid JSON: %w", err)
	}
	return nil
}

