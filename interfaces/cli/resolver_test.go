package cli

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

func TestTenantResolver(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})
	mockService.AddTenant(&core.Tenant{
		ID:       "inactive-id",
		Name:     "inactive-tenant",
		IsActive: false,
	})

	resolver := NewTenantResolver(mockService, "TEST_TENANT")
	ctx := context.Background()

	// Test ResolveTenant
	resolvedCtx, err := resolver.ResolveTenant(ctx, "test-tenant")
	assert.NoError(t, err)

	tenant, ok := tenantcontext.GetTenant(resolvedCtx)
	assert.True(t, ok)
	assert.Equal(t, "test-tenant", tenant.Name)

	// Test tenant not found
	_, err = resolver.ResolveTenant(ctx, "unknown-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Test inactive tenant
	_, err = resolver.ResolveTenant(ctx, "inactive-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantInactiveError{}, err)
}

func TestTenantResolver_ResolveTenantFromArgs(t *testing.T) {
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})

	resolver := NewTenantResolver(mockService, "TEST_TENANT")
	ctx := context.Background()

	// Test ResolveTenantFromArgs with --tenant=name format
	args := []string{"--tenant=test-tenant"}
	resolvedCtx, err := resolver.ResolveTenantFromArgs(ctx, args)
	assert.NoError(t, err)

	tenant, ok := tenantcontext.GetTenant(resolvedCtx)
	assert.True(t, ok)
	assert.Equal(t, "test-tenant", tenant.Name)

	// Test ResolveTenantFromArgs with --tenant name format
	args = []string{"--tenant", "test-tenant"}
	resolvedCtx, err = resolver.ResolveTenantFromArgs(ctx, args)
	assert.NoError(t, err)

	tenant, ok = tenantcontext.GetTenant(resolvedCtx)
	assert.True(t, ok)
	assert.Equal(t, "test-tenant", tenant.Name)

	// Test no tenant argument
	args = []string{"--other-arg", "value"}
	_, err = resolver.ResolveTenantFromArgs(ctx, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant argument not found")

	// Test --tenant without value
	args = []string{"--tenant"}
	_, err = resolver.ResolveTenantFromArgs(ctx, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant argument not found")
}

func TestTenantResolver_ResolveTenantFromEnv(t *testing.T) {
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})

	resolver := NewTenantResolver(mockService, "TEST_TENANT")
	ctx := context.Background()

	// Test with environment variable set
	os.Setenv("TEST_TENANT", "test-tenant")
	defer os.Unsetenv("TEST_TENANT")

	resolvedCtx, err := resolver.ResolveTenantFromEnv(ctx)
	assert.NoError(t, err)

	tenant, ok := tenantcontext.GetTenant(resolvedCtx)
	assert.True(t, ok)
	assert.Equal(t, "test-tenant", tenant.Name)

	// Test with environment variable not set
	os.Unsetenv("TEST_TENANT")
	_, err = resolver.ResolveTenantFromEnv(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable TEST_TENANT not set")
}

func TestTenantResolver_ForEachTenant(t *testing.T) {
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id-2",
		Name:     "test-tenant-2",
		IsActive: true,
	})
	mockService.AddTenant(&core.Tenant{
		ID:       "inactive-id",
		Name:     "inactive-tenant",
		IsActive: false,
	})

	resolver := NewTenantResolver(mockService, "TEST_TENANT")
	ctx := context.Background()

	// Test ForEachTenant
	var processedTenants []string
	err := resolver.ForEachTenant(ctx, func(ctx context.Context) error {
		tenant, _ := tenantcontext.GetTenant(ctx)
		processedTenants = append(processedTenants, tenant.Name)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, processedTenants, 2) // Only active tenants
	assert.Contains(t, processedTenants, "test-tenant")
	assert.Contains(t, processedTenants, "test-tenant-2")
	assert.NotContains(t, processedTenants, "inactive-tenant")
}

func TestTenantResolver_WithTenant(t *testing.T) {
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})

	resolver := NewTenantResolver(mockService, "TEST_TENANT")
	ctx := context.Background()

	// Test WithTenant
	var processedTenant string
	err := resolver.WithTenant(ctx, "test-tenant", func(ctx context.Context) error {
		tenant, _ := tenantcontext.GetTenant(ctx)
		processedTenant = tenant.Name
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "test-tenant", processedTenant)

	// Test WithTenant with non-existent tenant
	err = resolver.WithTenant(ctx, "unknown-tenant", func(ctx context.Context) error {
		return nil
	})

	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestNewTenantResolver(t *testing.T) {
	mockService := NewMockTenantService()

	// Test with custom env var name
	resolver := NewTenantResolver(mockService, "CUSTOM_TENANT")
	assert.Equal(t, "CUSTOM_TENANT", resolver.envVarName)

	// Test with empty env var name (should default to TENANT_NAME)
	resolver = NewTenantResolver(mockService, "")
	assert.Equal(t, "TENANT_NAME", resolver.envVarName)
}
