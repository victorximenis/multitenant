package mongodb

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/victorximenis/multitenant/core"
	"go.mongodb.org/mongo-driver/bson"
)

func setupTestMongoDB(t *testing.T) (*TenantRepository, func()) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx,
		"mongo:5",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections"),
		),
	)
	require.NoError(t, err)

	// Get connection details
	connectionString, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repository
	repo, err := NewTenantRepository(ctx, connectionString)
	require.NoError(t, err)

	cleanup := func() {
		mongoContainer.Terminate(ctx)
	}

	return repo, cleanup
}

func TestMongoDBTenantRepository(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test creating a tenant
	tenant := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "test-tenant",
		IsActive: true,
		Metadata: map[string]interface{}{"plan": "pro"},
		Datasources: []core.Datasource{
			{
				ID:       uuid.New().String(),
				TenantID: "",
				DSN:      "mongodb://user:pass@host:port/db",
				Role:     "rw",
				PoolSize: 10,
			},
		},
	}
	// Set TenantID after creating the tenant ID
	tenant.Datasources[0].TenantID = tenant.ID

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Test retrieving the tenant
	retrieved, err := repo.GetByName(ctx, "test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)
	assert.Len(t, retrieved.Datasources, 1)

	// Test inactive tenant
	tenant.IsActive = false
	err = repo.Update(ctx, tenant)
	assert.NoError(t, err)

	_, err = repo.GetByName(ctx, "test-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantInactiveError{}, err)
}

func TestTenantRepository_GetByNameNotFound(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test getting non-existent tenant
	_, err := repo.GetByName(ctx, "non-existent")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_List(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple tenants
	tenant1 := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "tenant-1",
		IsActive: true,
		Metadata: map[string]interface{}{"plan": "basic"},
	}

	tenant2 := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "tenant-2",
		IsActive: true,
		Metadata: map[string]interface{}{"plan": "pro"},
	}

	err := repo.Create(ctx, tenant1)
	assert.NoError(t, err)

	err = repo.Create(ctx, tenant2)
	assert.NoError(t, err)

	// Test listing all tenants
	tenants, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tenants, 2)

	// Verify tenant names
	names := make(map[string]bool)
	for _, tenant := range tenants {
		names[tenant.Name] = true
	}
	assert.True(t, names["tenant-1"])
	assert.True(t, names["tenant-2"])
}

func TestTenantRepository_Update(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant
	tenant := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "test-tenant",
		IsActive: true,
		Metadata: map[string]interface{}{"plan": "basic"},
	}

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Update the tenant
	tenant.Metadata["plan"] = "pro"
	tenant.IsActive = false

	err = repo.Update(ctx, tenant)
	assert.NoError(t, err)

	// Verify the update
	_, err = repo.GetByName(ctx, "test-tenant")
	assert.Error(t, err) // Should error because tenant is inactive
	assert.IsType(t, core.TenantInactiveError{}, err)
}

func TestTenantRepository_UpdateNotFound(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to update non-existent tenant
	tenant := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "non-existent",
		IsActive: true,
	}

	err := repo.Update(ctx, tenant)
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant
	tenant := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "test-tenant",
		IsActive: true,
	}

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Delete the tenant
	err = repo.Delete(ctx, tenant.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByName(ctx, "test-tenant")
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_DeleteNotFound(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Try to delete non-existent tenant
	err := repo.Delete(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}

func TestTenantRepository_CreateDuplicate(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant
	tenant1 := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "test-tenant",
		IsActive: true,
	}

	err := repo.Create(ctx, tenant1)
	assert.NoError(t, err)

	// Try to create another tenant with the same name
	tenant2 := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "test-tenant", // Same name
		IsActive: true,
	}

	err = repo.Create(ctx, tenant2)
	assert.Error(t, err) // Should fail due to unique index on name
}

func TestTenantRepository_EmbeddedDatasources(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant with multiple datasources
	tenant := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "multi-ds-tenant",
		IsActive: true,
		Datasources: []core.Datasource{
			{
				ID:       uuid.New().String(),
				TenantID: "",
				DSN:      "postgres://user:pass@host1:5432/db1",
				Role:     "read",
				PoolSize: 5,
			},
			{
				ID:       uuid.New().String(),
				TenantID: "",
				DSN:      "postgres://user:pass@host2:5432/db2",
				Role:     "rw",
				PoolSize: 10,
			},
		},
	}

	// Set TenantID for all datasources
	for i := range tenant.Datasources {
		tenant.Datasources[i].TenantID = tenant.ID
	}

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Retrieve and verify embedded datasources
	retrieved, err := repo.GetByName(ctx, "multi-ds-tenant")
	assert.NoError(t, err)
	assert.Len(t, retrieved.Datasources, 2)

	// Verify datasource details
	assert.Equal(t, "read", retrieved.Datasources[0].Role)
	assert.Equal(t, "rw", retrieved.Datasources[1].Role)
	assert.Equal(t, 5, retrieved.Datasources[0].PoolSize)
	assert.Equal(t, 10, retrieved.Datasources[1].PoolSize)
}

func TestTenantRepository_IndexCreation(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Get index information
	cursor, err := repo.collection.Indexes().List(ctx)
	require.NoError(t, err)

	var indexes []bson.M
	err = cursor.All(ctx, &indexes)
	require.NoError(t, err)

	// Verify that indexes were created
	indexNames := make(map[string]bool)
	for _, index := range indexes {
		if name, ok := index["name"].(string); ok {
			indexNames[name] = true
		}
	}

	// Check for expected indexes
	assert.True(t, indexNames["name_1"], "Unique index on 'name' should exist")
	assert.True(t, indexNames["datasources.id_1"], "Index on 'datasources.id' should exist")
	assert.True(t, indexNames["id_1"], "Index on 'id' should exist")
	assert.True(t, indexNames["is_active_1"], "Index on 'is_active' should exist")
}

func TestTenantRepository_UniqueConstraint(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create first tenant
	tenant1 := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "unique-test",
		IsActive: true,
	}

	err := repo.Create(ctx, tenant1)
	assert.NoError(t, err)

	// Try to create second tenant with same name
	tenant2 := &core.Tenant{
		ID:       uuid.New().String(),
		Name:     "unique-test", // Same name
		IsActive: true,
	}

	err = repo.Create(ctx, tenant2)
	assert.Error(t, err, "Should fail due to unique constraint on name")
}

func TestTenantRepository_AddDatasource(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant
	tenant := &core.Tenant{
		ID:          uuid.New().String(),
		Name:        "test-tenant",
		IsActive:    true,
		Datasources: []core.Datasource{},
	}

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Add a datasource
	newDatasource := core.Datasource{
		ID:       uuid.New().String(),
		TenantID: tenant.ID,
		DSN:      "postgres://user:pass@host:5432/db",
		Role:     "rw",
		PoolSize: 10,
	}

	err = repo.AddDatasource(ctx, tenant.ID, newDatasource)
	assert.NoError(t, err)

	// Verify the datasource was added
	retrieved, err := repo.GetByName(ctx, "test-tenant")
	assert.NoError(t, err)
	assert.Len(t, retrieved.Datasources, 1)
	assert.Equal(t, newDatasource.ID, retrieved.Datasources[0].ID)
	assert.Equal(t, newDatasource.DSN, retrieved.Datasources[0].DSN)
}

func TestTenantRepository_RemoveDatasource(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant with a datasource
	datasource := core.Datasource{
		ID:       uuid.New().String(),
		TenantID: "",
		DSN:      "postgres://user:pass@host:5432/db",
		Role:     "rw",
		PoolSize: 10,
	}

	tenant := &core.Tenant{
		ID:          uuid.New().String(),
		Name:        "test-tenant",
		IsActive:    true,
		Datasources: []core.Datasource{datasource},
	}
	tenant.Datasources[0].TenantID = tenant.ID

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Remove the datasource
	err = repo.RemoveDatasource(ctx, tenant.ID, datasource.ID)
	assert.NoError(t, err)

	// Verify the datasource was removed
	retrieved, err := repo.GetByName(ctx, "test-tenant")
	assert.NoError(t, err)
	assert.Len(t, retrieved.Datasources, 0)
}

func TestTenantRepository_UpdateDatasource(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant with a datasource
	datasource := core.Datasource{
		ID:       uuid.New().String(),
		TenantID: "",
		DSN:      "postgres://user:pass@host:5432/db",
		Role:     "read",
		PoolSize: 5,
	}

	tenant := &core.Tenant{
		ID:          uuid.New().String(),
		Name:        "test-tenant",
		IsActive:    true,
		Datasources: []core.Datasource{datasource},
	}
	tenant.Datasources[0].TenantID = tenant.ID

	err := repo.Create(ctx, tenant)
	assert.NoError(t, err)

	// Update the datasource
	updatedDatasource := datasource
	updatedDatasource.Role = "rw"
	updatedDatasource.PoolSize = 15
	updatedDatasource.DSN = "postgres://newuser:newpass@newhost:5432/newdb"

	err = repo.UpdateDatasource(ctx, tenant.ID, updatedDatasource)
	assert.NoError(t, err)

	// Verify the datasource was updated
	retrieved, err := repo.GetByName(ctx, "test-tenant")
	assert.NoError(t, err)
	assert.Len(t, retrieved.Datasources, 1)
	assert.Equal(t, "rw", retrieved.Datasources[0].Role)
	assert.Equal(t, 15, retrieved.Datasources[0].PoolSize)
	assert.Equal(t, "postgres://newuser:newpass@newhost:5432/newdb", retrieved.Datasources[0].DSN)
}

func TestTenantRepository_DatasourceOperationsNotFound(t *testing.T) {
	repo, cleanup := setupTestMongoDB(t)
	defer cleanup()

	ctx := context.Background()

	nonExistentID := uuid.New().String()
	datasource := core.Datasource{
		ID:       uuid.New().String(),
		TenantID: nonExistentID,
		DSN:      "postgres://user:pass@host:5432/db",
		Role:     "rw",
		PoolSize: 10,
	}

	// Test AddDatasource with non-existent tenant
	err := repo.AddDatasource(ctx, nonExistentID, datasource)
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Test RemoveDatasource with non-existent tenant
	err = repo.RemoveDatasource(ctx, nonExistentID, datasource.ID)
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Test UpdateDatasource with non-existent tenant
	err = repo.UpdateDatasource(ctx, nonExistentID, datasource)
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)
}
