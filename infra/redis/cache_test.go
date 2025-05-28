package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/victorximenis/multitenant/core"
)

func setupTestRedis(t *testing.T) (*TenantCache, func()) {
	ctx := context.Background()

	// Start Redis container
	redisContainer, err := redis.Run(ctx,
		"redis:6",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	require.NoError(t, err)

	// Get connection details
	connectionString, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create cache
	cache, err := NewTenantCache(ctx, Config{
		RedisURL: connectionString,
		TTL:      1 * time.Second, // Short TTL for testing
	})
	require.NoError(t, err)

	cleanup := func() {
		redisContainer.Terminate(ctx)
	}

	return cache, cleanup
}

func TestTenantCache(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	// Test caching a tenant
	tenant := &core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
		Datasources: []core.Datasource{
			{
				ID:       "ds-id",
				TenantID: "test-id",
				DSN:      "postgres://user:pass@host:port/db",
				Role:     "rw",
				PoolSize: 10,
			},
		},
	}

	err := cache.Set(ctx, tenant, 0)
	assert.NoError(t, err)

	// Test retrieving from cache
	retrieved, err := cache.Get(ctx, "test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)
	assert.Len(t, retrieved.Datasources, 1)

	// Test TTL expiration
	time.Sleep(2 * time.Second)
	_, err = cache.Get(ctx, "test-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Test cache invalidation
	err = cache.Set(ctx, tenant, 10*time.Second)
	assert.NoError(t, err)

	err = cache.Delete(ctx, "test-tenant")
	assert.NoError(t, err)

	_, err = cache.Get(ctx, "test-tenant")
	assert.Error(t, err)
}

func TestTenantCache_GetNotFound(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	// Test getting non-existent tenant
	_, err := cache.Get(ctx, "non-existent")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantCache_SetNilTenant(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	// Test setting nil tenant
	err := cache.Set(ctx, nil, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant cannot be nil")
}

func TestTenantCache_DeleteAll(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple tenants
	tenant1 := &core.Tenant{
		ID:       "test-id-1",
		Name:     "test-tenant-1",
		IsActive: true,
	}
	tenant2 := &core.Tenant{
		ID:       "test-id-2",
		Name:     "test-tenant-2",
		IsActive: true,
	}

	err := cache.Set(ctx, tenant1, 10*time.Second)
	assert.NoError(t, err)

	err = cache.Set(ctx, tenant2, 10*time.Second)
	assert.NoError(t, err)

	// Verify both tenants exist
	_, err = cache.Get(ctx, "test-tenant-1")
	assert.NoError(t, err)

	_, err = cache.Get(ctx, "test-tenant-2")
	assert.NoError(t, err)

	// Delete all tenants
	err = cache.DeleteAll(ctx)
	assert.NoError(t, err)

	// Verify both tenants are gone
	_, err = cache.Get(ctx, "test-tenant-1")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	_, err = cache.Get(ctx, "test-tenant-2")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantCache_CustomTTL(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	tenant := &core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	}

	// Set with custom TTL
	customTTL := 500 * time.Millisecond
	err := cache.Set(ctx, tenant, customTTL)
	assert.NoError(t, err)

	// Should exist immediately
	_, err = cache.Get(ctx, "test-tenant")
	assert.NoError(t, err)

	// Wait for custom TTL to expire
	time.Sleep(600 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "test-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantCache_ConcurrentAccess(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	tenant := &core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	}

	// Test concurrent writes and reads
	done := make(chan bool, 10)

	// Start multiple goroutines writing
	for i := 0; i < 5; i++ {
		go func(id int) {
			tenantCopy := *tenant
			tenantCopy.Name = fmt.Sprintf("test-tenant-%d", id)
			tenantCopy.ID = fmt.Sprintf("test-id-%d", id)
			err := cache.Set(ctx, &tenantCopy, 10*time.Second)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Start multiple goroutines reading
	for i := 0; i < 5; i++ {
		go func(id int) {
			time.Sleep(100 * time.Millisecond) // Give writes time to complete
			tenantName := fmt.Sprintf("test-tenant-%d", id)
			_, err := cache.Get(ctx, tenantName)
			// May or may not find the tenant depending on timing, but should not panic
			if err != nil {
				assert.IsType(t, core.TenantNotFoundError{}, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
