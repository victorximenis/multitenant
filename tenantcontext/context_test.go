package tenantcontext

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenantContext(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := context.Background()

	// Test storing and retrieving
	ctx = WithTenant(ctx, tenant)
	retrieved, ok := GetTenant(ctx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)

	// Test tenant name retrieval
	name := GetCurrentTenantName(ctx)
	assert.Equal(t, "test-tenant", name)

	// Test tenant ID retrieval
	id := GetCurrentTenantID(ctx)
	assert.Equal(t, tenant.ID, id)

	// Test HasTenant
	assert.True(t, HasTenant(ctx))

	// Test empty context
	emptyCtx := context.Background()
	_, ok = GetTenant(emptyCtx)
	assert.False(t, ok)
	assert.Empty(t, GetCurrentTenantName(emptyCtx))
	assert.Empty(t, GetCurrentTenantID(emptyCtx))
	assert.False(t, HasTenant(emptyCtx))

	// Test MustGetTenant panic
	assert.Panics(t, func() {
		MustGetTenant(emptyCtx)
	})
}

func TestWithTenantNilContext(t *testing.T) {
	tenant := NewTestTenant("test-tenant")

	// Test with nil context - should create background context
	ctx := WithTenant(nil, tenant)
	retrieved, ok := GetTenant(ctx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)
}

func TestGetTenantNilContext(t *testing.T) {
	// Test with nil context
	tenant, ok := GetTenant(nil)
	assert.False(t, ok)
	assert.Nil(t, tenant)
}

func TestMustGetTenantSuccess(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	retrieved := MustGetTenant(ctx)
	assert.Equal(t, tenant, retrieved)
}

func TestConcurrentAccess(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	results := make(chan bool, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Create tenant with unique name
				tenant := NewTestTenant("tenant-" + string(rune(id)) + "-" + string(rune(j)))
				ctx := WithTenant(context.Background(), tenant)

				// Verify retrieval
				retrieved, ok := GetTenant(ctx)
				results <- ok && retrieved.Name == tenant.Name
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Check all operations succeeded
	successCount := 0
	for result := range results {
		if result {
			successCount++
		}
	}

	assert.Equal(t, numGoroutines*numOperations, successCount)
}

func TestContextPropagation(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	// Test context propagation through function calls
	result := executeWithContext(ctx)
	assert.Equal(t, "test-tenant", result)
}

func executeWithContext(ctx context.Context) string {
	return GetCurrentTenantName(ctx)
}

func TestMultipleTenants(t *testing.T) {
	tenant1 := NewTestTenant("tenant-1")
	tenant2 := NewTestTenant("tenant-2")

	ctx1 := WithTenant(context.Background(), tenant1)
	ctx2 := WithTenant(context.Background(), tenant2)

	// Verify each context has the correct tenant
	name1 := GetCurrentTenantName(ctx1)
	name2 := GetCurrentTenantName(ctx2)

	assert.Equal(t, "tenant-1", name1)
	assert.Equal(t, "tenant-2", name2)
	assert.NotEqual(t, name1, name2)
}

func TestContextOverwrite(t *testing.T) {
	tenant1 := NewTestTenant("tenant-1")
	tenant2 := NewTestTenant("tenant-2")

	ctx := WithTenant(context.Background(), tenant1)
	assert.Equal(t, "tenant-1", GetCurrentTenantName(ctx))

	// Overwrite with second tenant
	ctx = WithTenant(ctx, tenant2)
	assert.Equal(t, "tenant-2", GetCurrentTenantName(ctx))
}

func TestTestHelpers(t *testing.T) {
	// Test NewTestContext
	ctx := NewTestContext()
	assert.True(t, HasTenant(ctx))
	assert.Equal(t, "test-tenant", GetCurrentTenantName(ctx))

	// Test NewTestContextWithName
	ctx = NewTestContextWithName("custom-tenant")
	assert.Equal(t, "custom-tenant", GetCurrentTenantName(ctx))

	// Test AssertTenantInContext
	assert.True(t, AssertTenantInContext(ctx, "custom-tenant"))
	assert.False(t, AssertTenantInContext(ctx, "wrong-tenant"))

	// Test inactive tenant
	ctx = NewTestContextWithInactiveTenant("inactive-tenant")
	tenant, ok := GetTenant(ctx)
	assert.True(t, ok)
	assert.False(t, tenant.IsActive)
	assert.Equal(t, "inactive-tenant", tenant.Name)
}

func TestCreateTestTenantWithDatasources(t *testing.T) {
	tenant := CreateTestTenantWithDatasources("test-tenant", 3)

	assert.Equal(t, "test-tenant", tenant.Name)
	assert.Len(t, tenant.Datasources, 3)

	// Verify all datasources belong to the tenant
	for _, ds := range tenant.Datasources {
		assert.Equal(t, tenant.ID, ds.TenantID)
		assert.Equal(t, "rw", ds.Role)
		assert.Equal(t, 10, ds.PoolSize)
	}
}
