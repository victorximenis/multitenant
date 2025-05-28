package core

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multitenant system
type Tenant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
	Datasources []Datasource           `json:"datasources"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Validate validates the tenant data
func (t *Tenant) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("tenant name cannot be empty")
	}

	if t.ID == "" {
		return errors.New("tenant ID cannot be empty")
	}

	// Validate UUID format for ID
	if _, err := uuid.Parse(t.ID); err != nil {
		return errors.New("tenant ID must be a valid UUID")
	}

	// Validate datasources
	for i, ds := range t.Datasources {
		if err := ds.Validate(); err != nil {
			return errors.New("datasource " + string(rune(i)) + " validation failed: " + err.Error())
		}
	}

	return nil
}

// NewTenant creates a new tenant with generated ID and timestamps
func NewTenant(name string) *Tenant {
	now := time.Now()
	return &Tenant{
		ID:          uuid.New().String(),
		Name:        name,
		IsActive:    true,
		Metadata:    make(map[string]interface{}),
		Datasources: make([]Datasource, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Datasource represents a database connection configuration for a tenant
type Datasource struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	DSN       string                 `json:"dsn"`
	Role      string                 `json:"role"` // read, write, rw
	PoolSize  int                    `json:"pool_size"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Validate validates the datasource data
func (d *Datasource) Validate() error {
	if d.ID == "" {
		return errors.New("datasource ID cannot be empty")
	}

	if d.TenantID == "" {
		return errors.New("datasource tenant ID cannot be empty")
	}

	if strings.TrimSpace(d.DSN) == "" {
		return errors.New("datasource DSN cannot be empty")
	}

	validRoles := map[string]bool{
		"read":  true,
		"write": true,
		"rw":    true,
	}

	if !validRoles[d.Role] {
		return errors.New("datasource role must be one of: read, write, rw")
	}

	if d.PoolSize < 1 {
		return errors.New("datasource pool size must be greater than 0")
	}

	// Validate UUID format for IDs
	if _, err := uuid.Parse(d.ID); err != nil {
		return errors.New("datasource ID must be a valid UUID")
	}

	if _, err := uuid.Parse(d.TenantID); err != nil {
		return errors.New("datasource tenant ID must be a valid UUID")
	}

	return nil
}

// NewDatasource creates a new datasource with generated ID and timestamps
func NewDatasource(tenantID, dsn, role string, poolSize int) *Datasource {
	now := time.Now()
	return &Datasource{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		DSN:       dsn,
		Role:      role,
		PoolSize:  poolSize,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
