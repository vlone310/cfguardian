package entities

import (
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// Project represents a project entity for multi-tenancy
type Project struct {
	id          string
	name        string
	apiKey      valueobjects.APIKey
	ownerUserID string
	createdAt   time.Time
	updatedAt   time.Time
}

// NewProject creates a new Project entity
func NewProject(id, name string, apiKey valueobjects.APIKey, ownerUserID string) *Project {
	now := time.Now()
	return &Project{
		id:          id,
		name:        name,
		apiKey:      apiKey,
		ownerUserID: ownerUserID,
		createdAt:   now,
		updatedAt:   now,
	}
}

// ReconstructProject reconstructs a Project from persistence layer
func ReconstructProject(id, name string, apiKey valueobjects.APIKey, ownerUserID string, createdAt, updatedAt time.Time) *Project {
	return &Project{
		id:          id,
		name:        name,
		apiKey:      apiKey,
		ownerUserID: ownerUserID,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ID returns the project ID
func (p *Project) ID() string {
	return p.id
}

// Name returns the project name
func (p *Project) Name() string {
	return p.name
}

// APIKey returns the project's API key
func (p *Project) APIKey() valueobjects.APIKey {
	return p.apiKey
}

// OwnerUserID returns the owner user ID
func (p *Project) OwnerUserID() string {
	return p.ownerUserID
}

// CreatedAt returns the creation timestamp
func (p *Project) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update timestamp
func (p *Project) UpdatedAt() time.Time {
	return p.updatedAt
}

// UpdateName updates the project name
func (p *Project) UpdateName(name string) {
	p.name = name
	p.updatedAt = time.Now()
}

// RegenerateAPIKey updates the project's API key
func (p *Project) RegenerateAPIKey(newAPIKey valueobjects.APIKey) {
	p.apiKey = newAPIKey
	p.updatedAt = time.Now()
}

// IsOwnedBy checks if the project is owned by a specific user
func (p *Project) IsOwnedBy(userID string) bool {
	return p.ownerUserID == userID
}

// Equals checks if two projects are the same entity (by ID)
func (p *Project) Equals(other *Project) bool {
	if other == nil {
		return false
	}
	return p.id == other.id
}

