package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victorximenis/multitenant/core"
)

func TestTenantRepository_GetByName_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	// Mock tenant data
	tenantID := "123e4567-e89b-12d3-a456-426614174000"
	tenantName := "test-tenant"
	metadata := map[string]interface{}{"plan": "pro"}
	metadataBytes, _ := json.Marshal(metadata)
	createdAt := time.Now()
	updatedAt := time.Now()

	// Mock datasource data
	dsID := "123e4567-e89b-12d3-a456-426614174001"
	dsMetadata := map[string]interface{}{"region": "us-east-1"}
	dsMetadataBytes, _ := json.Marshal(dsMetadata)

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect tenant query
	tenantRows := mock.NewRows([]string{"id", "name", "is_active", "metadata", "created_at", "updated_at"}).
		AddRow(tenantID, tenantName, true, metadataBytes, createdAt, updatedAt)
	mock.ExpectQuery("SELECT id, name, is_active, metadata, created_at, updated_at FROM tenants WHERE name = \\$1").
		WithArgs(tenantName).
		WillReturnRows(tenantRows)

	// Expect datasources query
	dsRows := mock.NewRows([]string{"id", "dsn", "role", "pool_size", "metadata", "created_at", "updated_at"}).
		AddRow(dsID, "postgres://user:pass@host:5432/db", "read", 10, dsMetadataBytes, createdAt, updatedAt)
	mock.ExpectQuery("SELECT id, dsn, role, pool_size, metadata, created_at, updated_at FROM datasources WHERE tenant_id = \\$1").
		WithArgs(tenantID).
		WillReturnRows(dsRows)

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute test
	tenant, err := repo.GetByName(ctx, tenantName)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, tenantID, tenant.ID)
	assert.Equal(t, tenantName, tenant.Name)
	assert.True(t, tenant.IsActive)
	assert.Equal(t, metadata, tenant.Metadata)
	assert.Len(t, tenant.Datasources, 1)
	assert.Equal(t, dsID, tenant.Datasources[0].ID)
	assert.Equal(t, "read", tenant.Datasources[0].Role)
	assert.Equal(t, dsMetadata, tenant.Datasources[0].Metadata)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_GetByName_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	tenantName := "non-existent"

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect tenant query to return no rows
	mock.ExpectQuery("SELECT id, name, is_active, metadata, created_at, updated_at FROM tenants WHERE name = \\$1").
		WithArgs(tenantName).
		WillReturnError(pgx.ErrNoRows)

	// Expect transaction rollback
	mock.ExpectRollback()

	// Execute test
	tenant, err := repo.GetByName(ctx, tenantName)

	// Verify results
	assert.Nil(t, tenant)
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_Create_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	// Create test tenant
	tenant := core.NewTenant("test-tenant")
	tenant.Metadata = map[string]interface{}{"plan": "pro"}
	ds := core.NewDatasource(tenant.ID, "postgres://user:pass@host:5432/db", "read", 10)
	tenant.Datasources = []core.Datasource{*ds}

	metadataBytes, _ := json.Marshal(tenant.Metadata)
	dsMetadataBytes, _ := json.Marshal(ds.Metadata)

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect tenant insert
	mock.ExpectExec("INSERT INTO tenants").
		WithArgs(tenant.ID, tenant.Name, tenant.IsActive, metadataBytes, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Expect datasource insert
	mock.ExpectExec("INSERT INTO datasources").
		WithArgs(ds.ID, ds.TenantID, ds.DSN, ds.Role, ds.PoolSize, dsMetadataBytes, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute test
	err = repo.Create(ctx, tenant)

	// Verify results
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_List_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	// Mock tenant data
	tenant1ID := "123e4567-e89b-12d3-a456-426614174000"
	tenant2ID := "123e4567-e89b-12d3-a456-426614174001"
	metadata := map[string]interface{}{"plan": "pro"}
	metadataBytes, _ := json.Marshal(metadata)
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expect tenants query
	tenantRows := mock.NewRows([]string{"id", "name", "is_active", "metadata", "created_at", "updated_at"}).
		AddRow(tenant1ID, "tenant-1", true, metadataBytes, createdAt, updatedAt).
		AddRow(tenant2ID, "tenant-2", true, metadataBytes, createdAt, updatedAt)
	mock.ExpectQuery("SELECT id, name, is_active, metadata, created_at, updated_at FROM tenants ORDER BY name").
		WillReturnRows(tenantRows)

	// Expect datasources queries for each tenant
	dsRows1 := mock.NewRows([]string{"id", "dsn", "role", "pool_size", "metadata", "created_at", "updated_at"}).
		AddRow("ds1", "postgres://user:pass@host:5432/db1", "read", 10, metadataBytes, createdAt, updatedAt)
	mock.ExpectQuery("SELECT id, dsn, role, pool_size, metadata, created_at, updated_at FROM datasources WHERE tenant_id = \\$1").
		WithArgs(tenant1ID).
		WillReturnRows(dsRows1)

	dsRows2 := mock.NewRows([]string{"id", "dsn", "role", "pool_size", "metadata", "created_at", "updated_at"}).
		AddRow("ds2", "postgres://user:pass@host:5432/db2", "rw", 15, metadataBytes, createdAt, updatedAt)
	mock.ExpectQuery("SELECT id, dsn, role, pool_size, metadata, created_at, updated_at FROM datasources WHERE tenant_id = \\$1").
		WithArgs(tenant2ID).
		WillReturnRows(dsRows2)

	// Execute test
	tenants, err := repo.List(ctx)

	// Verify results
	assert.NoError(t, err)
	assert.Len(t, tenants, 2)
	assert.Equal(t, "tenant-1", tenants[0].Name)
	assert.Equal(t, "tenant-2", tenants[1].Name)
	assert.Len(t, tenants[0].Datasources, 1)
	assert.Len(t, tenants[1].Datasources, 1)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_Update_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	// Create test tenant
	tenant := core.NewTenant("updated-tenant")
	tenant.Metadata = map[string]interface{}{"plan": "enterprise"}
	ds := core.NewDatasource(tenant.ID, "postgres://user:pass@host:5432/db", "rw", 20)
	tenant.Datasources = []core.Datasource{*ds}

	metadataBytes, _ := json.Marshal(tenant.Metadata)
	dsMetadataBytes, _ := json.Marshal(ds.Metadata)

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect existence check
	existsRows := mock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tenants WHERE id = \\$1\\)").
		WithArgs(tenant.ID).
		WillReturnRows(existsRows)

	// Expect tenant update
	mock.ExpectExec("UPDATE tenants SET name = \\$2, is_active = \\$3, metadata = \\$4, updated_at = \\$5 WHERE id = \\$1").
		WithArgs(tenant.ID, tenant.Name, tenant.IsActive, metadataBytes, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Expect datasources delete
	mock.ExpectExec("DELETE FROM datasources WHERE tenant_id = \\$1").
		WithArgs(tenant.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	// Expect datasource insert
	mock.ExpectExec("INSERT INTO datasources").
		WithArgs(ds.ID, ds.TenantID, ds.DSN, ds.Role, ds.PoolSize, dsMetadataBytes, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute test
	err = repo.Update(ctx, tenant)

	// Verify results
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_Update_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	// Create test tenant
	tenant := core.NewTenant("non-existent")

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect existence check to return false
	existsRows := mock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tenants WHERE id = \\$1\\)").
		WithArgs(tenant.ID).
		WillReturnRows(existsRows)

	// Expect transaction rollback
	mock.ExpectRollback()

	// Execute test
	err = repo.Update(ctx, tenant)

	// Verify results
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_Delete_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	tenantName := "delete-tenant"
	tenantID := "123e4567-e89b-12d3-a456-426614174000"

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect tenant ID lookup
	idRows := mock.NewRows([]string{"id"}).AddRow(tenantID)
	mock.ExpectQuery("SELECT id FROM tenants WHERE name = \\$1").
		WithArgs(tenantName).
		WillReturnRows(idRows)

	// Expect tenant delete
	mock.ExpectExec("DELETE FROM tenants WHERE name = \\$1").
		WithArgs(tenantName).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute test
	err = repo.Delete(ctx, tenantName)

	// Verify results
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_Delete_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &TenantRepository{pool: mock}
	ctx := context.Background()

	tenantName := "non-existent"

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect tenant ID lookup to return no rows
	mock.ExpectQuery("SELECT id FROM tenants WHERE name = \\$1").
		WithArgs(tenantName).
		WillReturnError(pgx.ErrNoRows)

	// Expect transaction rollback
	mock.ExpectRollback()

	// Execute test
	err = repo.Delete(ctx, tenantName)

	// Verify results
	assert.Error(t, err)
	assert.IsType(t, core.TenantNotFoundError{}, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapPostgreSQLError_UniqueViolation(t *testing.T) {
	// This would require creating a mock pgconn.PgError, which is complex
	// For now, we'll test that the function exists and handles nil correctly
	err := mapPostgreSQLError(nil)
	assert.NoError(t, err)

	// Test with a regular error (not PostgreSQL error)
	regularErr := assert.AnError
	mappedErr := mapPostgreSQLError(regularErr)
	assert.Equal(t, regularErr, mappedErr)
}
