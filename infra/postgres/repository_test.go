package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/victorximenis/multitenant/core"
)

// setupTestDB creates a PostgreSQL container and returns a repository instance
func setupTestDB(t *testing.T) (*TenantRepository, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:14",
		postgres.WithDatabase("multitenant_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	// Get connection details
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create repository
	repo, err := NewTenantRepository(ctx, connStr)
	require.NoError(t, err)

	cleanup := func() {
		if repo != nil {
			repo.Close()
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return repo, cleanup
}

func TestTenantRepository_Create(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test tenant
	tenant := core.NewTenant("test-tenant")
	tenant.Metadata = map[string]interface{}{
		"plan":      "pro",
		"max_users": 100,
		"features":  []string{"analytics", "api"},
	}

	// Add datasources
	ds1 := core.NewDatasource(tenant.ID, "postgres://user:pass@host:5432/db1", "read", 5)
	ds1.Metadata = map[string]interface{}{"region": "us-east-1"}

	ds2 := core.NewDatasource(tenant.ID, "postgres://user:pass@host:5432/db2", "rw", 10)
	ds2.Metadata = map[string]interface{}{"region": "us-west-1"}

	tenant.Datasources = []core.Datasource{*ds1, *ds2}

	// Test creating the tenant
	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Verify the tenant was created
	retrieved, err := repo.GetByName(ctx, "test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)
	assert.Equal(t, tenant.IsActive, retrieved.IsActive)
	assert.Equal(t, tenant.Metadata, retrieved.Metadata)
	assert.Len(t, retrieved.Datasources, 2)

	// Verify datasources
	assert.Equal(t, ds1.ID, retrieved.Datasources[0].ID)
	assert.Equal(t, ds1.DSN, retrieved.Datasources[0].DSN)
	assert.Equal(t, ds1.Role, retrieved.Datasources[0].Role)
	assert.Equal(t, ds1.PoolSize, retrieved.Datasources[0].PoolSize)
	assert.Equal(t, ds1.Metadata, retrieved.Datasources[0].Metadata)
}

func TestTenantRepository_CreateDuplicate(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create first tenant
	tenant1 := core.NewTenant("duplicate-tenant")
	err := repo.Create(ctx, tenant1)
	assert.NoError(t, err)

	// Try to create tenant with same name
	tenant2 := core.NewTenant("duplicate-tenant")
	err = repo.Create(ctx, tenant2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant name already exists")
}

func TestTenantRepository_GetByName(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test getting non-existent tenant
	_, err := repo.GetByName(ctx, "non-existent")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Create and retrieve tenant
	tenant := core.NewTenant("get-test-tenant")
	err = repo.Create(ctx, tenant)
	assert.NoError(t, err)

	retrieved, err := repo.GetByName(ctx, "get-test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)
}

func TestTenantRepository_List(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test empty list
	tenants, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Empty(t, tenants)

	// Create multiple tenants
	tenant1 := core.NewTenant("tenant-1")
	tenant2 := core.NewTenant("tenant-2")
	tenant3 := core.NewTenant("tenant-3")

	err = repo.Create(ctx, tenant1)
	assert.NoError(t, err)
	err = repo.Create(ctx, tenant2)
	assert.NoError(t, err)
	err = repo.Create(ctx, tenant3)
	assert.NoError(t, err)

	// List all tenants
	tenants, err = repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tenants, 3)

	// Verify they are sorted by name
	assert.Equal(t, "tenant-1", tenants[0].Name)
	assert.Equal(t, "tenant-2", tenants[1].Name)
	assert.Equal(t, "tenant-3", tenants[2].Name)
}

func TestTenantRepository_Update(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create initial tenant
	tenant := core.NewTenant("update-tenant")
	tenant.Metadata = map[string]interface{}{"version": 1}
	ds := core.NewDatasource(tenant.ID, "postgres://old:pass@host:5432/db", "read", 5)
	tenant.Datasources = []core.Datasource{*ds}

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Update tenant
	tenant.Name = "updated-tenant"
	tenant.IsActive = false
	tenant.Metadata = map[string]interface{}{"version": 2, "updated": true}

	// Replace datasources
	newDS := core.NewDatasource(tenant.ID, "postgres://new:pass@host:5432/db", "rw", 15)
	tenant.Datasources = []core.Datasource{*newDS}

	err = repo.Update(ctx, tenant)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByName(ctx, "updated-tenant")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, "updated-tenant", retrieved.Name)
	assert.False(t, retrieved.IsActive)
	assert.Equal(t, tenant.Metadata, retrieved.Metadata)
	assert.Len(t, retrieved.Datasources, 1)
	assert.Equal(t, newDS.DSN, retrieved.Datasources[0].DSN)
	assert.Equal(t, "rw", retrieved.Datasources[0].Role)
	assert.Equal(t, 15, retrieved.Datasources[0].PoolSize)
}

func TestTenantRepository_UpdateNonExistent(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to update non-existent tenant
	tenant := core.NewTenant("non-existent")
	err := repo.Update(ctx, tenant)
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant with datasources
	tenant := core.NewTenant("delete-tenant")
	ds := core.NewDatasource(tenant.ID, "postgres://user:pass@host:5432/db", "read", 5)
	tenant.Datasources = []core.Datasource{*ds}

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Verify tenant exists
	_, err = repo.GetByName(ctx, "delete-tenant")
	assert.NoError(t, err)

	// Delete tenant
	err = repo.Delete(ctx, "delete-tenant")
	assert.NoError(t, err)

	// Verify tenant is deleted
	_, err = repo.GetByName(ctx, "delete-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_DeleteNonExistent(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to delete non-existent tenant
	err := repo.Delete(ctx, "non-existent")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_TransactionRollback(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant with invalid datasource to trigger rollback
	tenant := core.NewTenant("rollback-tenant")

	// Create datasource with invalid role to trigger constraint violation
	invalidDS := &core.Datasource{
		ID:       "invalid-id", // This will fail UUID validation
		TenantID: tenant.ID,
		DSN:      "postgres://user:pass@host:5432/db",
		Role:     "read",
		PoolSize: 5,
		Metadata: make(map[string]interface{}),
	}
	tenant.Datasources = []core.Datasource{*invalidDS}

	// This should fail due to invalid datasource ID
	err := repo.Create(ctx, tenant)
	assert.Error(t, err)

	// Verify tenant was not created (transaction rolled back)
	_, err = repo.GetByName(ctx, "rollback-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_ConcurrentAccess(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test concurrent creation of different tenants
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			tenant := core.NewTenant(fmt.Sprintf("concurrent-tenant-%d", id))
			err := repo.Create(ctx, tenant)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify all tenants were created
	tenants, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tenants, 10)
}

func TestTenantRepository_SchemaSetup(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container without schema setup
	pgContainer, err := postgres.Run(ctx,
		"postgres:14",
		postgres.WithDatabase("schema_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create repository (this should set up the schema)
	repo, err := NewTenantRepository(ctx, connStr)
	require.NoError(t, err)
	defer repo.Close()

	// Test that we can create a tenant (schema is working)
	tenant := core.NewTenant("schema-test-tenant")
	err = repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Verify tenant was created
	retrieved, err := repo.GetByName(ctx, "schema-test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
}
