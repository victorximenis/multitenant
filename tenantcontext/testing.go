package tenantcontext

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/victorximenis/multitenant/core"
)

// NewTestTenant creates a test tenant with default values
func NewTestTenant(name string) *core.Tenant {
	if name == "" {
		name = "test-tenant"
	}

	return &core.Tenant{
		ID:          uuid.New().String(),
		Name:        name,
		IsActive:    true,
		Metadata:    make(map[string]interface{}),
		Datasources: []core.Datasource{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestContext creates a context with a test tenant
func NewTestContext() context.Context {
	tenant := NewTestTenant("test-tenant")
	return WithTenant(context.Background(), tenant)
}

// NewTestContextWithTenant creates a context with a specific test tenant
func NewTestContextWithTenant(tenant *core.Tenant) context.Context {
	return WithTenant(context.Background(), tenant)
}

// NewTestContextWithName creates a context with a test tenant with specific name
func NewTestContextWithName(name string) context.Context {
	tenant := NewTestTenant(name)
	return WithTenant(context.Background(), tenant)
}

// NewInactiveTenant creates a test tenant that is inactive
func NewInactiveTenant(name string) *core.Tenant {
	tenant := NewTestTenant(name)
	tenant.IsActive = false
	return tenant
}

// NewTestContextWithInactiveTenant creates a context with an inactive test tenant
func NewTestContextWithInactiveTenant(name string) context.Context {
	tenant := NewInactiveTenant(name)
	return WithTenant(context.Background(), tenant)
}

// AssertTenantInContext is a helper function to verify tenant presence in context
func AssertTenantInContext(ctx context.Context, expectedName string) bool {
	tenant, ok := GetTenant(ctx)
	if !ok {
		return false
	}

	return tenant.Name == expectedName
}

// CreateTestTenantWithDatasources creates a test tenant with datasources
func CreateTestTenantWithDatasources(name string, datasourceCount int) *core.Tenant {
	tenant := NewTestTenant(name)

	for i := 0; i < datasourceCount; i++ {
		datasource := core.NewDatasource(
			tenant.ID,
			"postgres://test:test@localhost/test",
			"rw",
			10,
		)
		tenant.Datasources = append(tenant.Datasources, *datasource)
	}

	return tenant
}
