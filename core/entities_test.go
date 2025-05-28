package core

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTenantValidation(t *testing.T) {
	tests := []struct {
		name    string
		tenant  *Tenant
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tenant",
			tenant: &Tenant{
				ID:          uuid.New().String(),
				Name:        "test-tenant",
				IsActive:    true,
				Metadata:    make(map[string]interface{}),
				Datasources: []Datasource{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tenant: &Tenant{
				ID:          uuid.New().String(),
				Name:        "",
				IsActive:    true,
				Metadata:    make(map[string]interface{}),
				Datasources: []Datasource{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "tenant name cannot be empty",
		},
		{
			name: "whitespace only name",
			tenant: &Tenant{
				ID:          uuid.New().String(),
				Name:        "   ",
				IsActive:    true,
				Metadata:    make(map[string]interface{}),
				Datasources: []Datasource{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "tenant name cannot be empty",
		},
		{
			name: "missing ID",
			tenant: &Tenant{
				ID:          "",
				Name:        "test-tenant",
				IsActive:    true,
				Metadata:    make(map[string]interface{}),
				Datasources: []Datasource{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "tenant ID cannot be empty",
		},
		{
			name: "invalid UUID format",
			tenant: &Tenant{
				ID:          "invalid-uuid",
				Name:        "test-tenant",
				IsActive:    true,
				Metadata:    make(map[string]interface{}),
				Datasources: []Datasource{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "tenant ID must be a valid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewTenant(t *testing.T) {
	name := "test-tenant"
	tenant := NewTenant(name)

	assert.NotNil(t, tenant)
	assert.Equal(t, name, tenant.Name)
	assert.True(t, tenant.IsActive)
	assert.NotEmpty(t, tenant.ID)
	assert.NotNil(t, tenant.Metadata)
	assert.NotNil(t, tenant.Datasources)
	assert.Equal(t, 0, len(tenant.Datasources))
	assert.False(t, tenant.CreatedAt.IsZero())
	assert.False(t, tenant.UpdatedAt.IsZero())

	// Validate that ID is a valid UUID
	_, err := uuid.Parse(tenant.ID)
	assert.NoError(t, err)

	// Validate that the tenant passes validation
	err = tenant.Validate()
	assert.NoError(t, err)
}

func TestDatasourceValidation(t *testing.T) {
	tenantID := uuid.New().String()

	tests := []struct {
		name       string
		datasource *Datasource
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid datasource",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: tenantID,
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			datasource: &Datasource{
				ID:       "",
				TenantID: tenantID,
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: "",
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource tenant ID cannot be empty",
		},
		{
			name: "missing DSN",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: tenantID,
				DSN:      "",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource DSN cannot be empty",
		},
		{
			name: "whitespace only DSN",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: tenantID,
				DSN:      "   ",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource DSN cannot be empty",
		},
		{
			name: "invalid role",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: tenantID,
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "invalid",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource role must be one of: read, write, rw",
		},
		{
			name: "zero pool size",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: tenantID,
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: 0,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource pool size must be greater than 0",
		},
		{
			name: "negative pool size",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: tenantID,
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: -1,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource pool size must be greater than 0",
		},
		{
			name: "invalid datasource ID UUID",
			datasource: &Datasource{
				ID:       "invalid-uuid",
				TenantID: tenantID,
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource ID must be a valid UUID",
		},
		{
			name: "invalid tenant ID UUID",
			datasource: &Datasource{
				ID:       uuid.New().String(),
				TenantID: "invalid-uuid",
				DSN:      "postgres://user:pass@localhost/db",
				Role:     "rw",
				PoolSize: 10,
				Metadata: make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "datasource tenant ID must be a valid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.datasource.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewDatasource(t *testing.T) {
	tenantID := uuid.New().String()
	dsn := "postgres://user:pass@localhost/db"
	role := "rw"
	poolSize := 10

	datasource := NewDatasource(tenantID, dsn, role, poolSize)

	assert.NotNil(t, datasource)
	assert.NotEmpty(t, datasource.ID)
	assert.Equal(t, tenantID, datasource.TenantID)
	assert.Equal(t, dsn, datasource.DSN)
	assert.Equal(t, role, datasource.Role)
	assert.Equal(t, poolSize, datasource.PoolSize)
	assert.NotNil(t, datasource.Metadata)

	// Validate that ID is a valid UUID
	_, err := uuid.Parse(datasource.ID)
	assert.NoError(t, err)

	// Validate that the datasource passes validation
	err = datasource.Validate()
	assert.NoError(t, err)
}

func TestDatasourceRoles(t *testing.T) {
	tenantID := uuid.New().String()
	validRoles := []string{"read", "write", "rw"}

	for _, role := range validRoles {
		t.Run("valid role: "+role, func(t *testing.T) {
			datasource := NewDatasource(tenantID, "postgres://user:pass@localhost/db", role, 10)
			err := datasource.Validate()
			assert.NoError(t, err)
		})
	}
}
