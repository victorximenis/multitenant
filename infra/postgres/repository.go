package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/victorximenis/multitenant/core"
)

// TenantRepository implements the core.TenantRepository interface using PostgreSQL
type TenantRepository struct {
	pool PoolInterface
}

// NewTenantRepository creates a new PostgreSQL tenant repository
func NewTenantRepository(ctx context.Context, dsn string) (*TenantRepository, error) {
	// Parse and configure the connection pool
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Configure pool settings for optimal performance
	config.MaxConns = 30
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	// Setup schema
	conn, err := pool.Acquire(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	defer conn.Release()

	if err := SetupSchema(ctx, conn.Conn()); err != nil {
		pool.Close()
		return nil, err
	}

	return &TenantRepository{pool: pool}, nil
}

// Close closes the connection pool
func (r *TenantRepository) Close() {
	r.pool.Close()
}

// GetByName retrieves a tenant by name with all its datasources
func (r *TenantRepository) GetByName(ctx context.Context, name string) (*core.Tenant, error) {
	tenant := &core.Tenant{}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get tenant
	row := tx.QueryRow(ctx, `
		SELECT id, name, is_active, metadata, created_at, updated_at 
		FROM tenants WHERE name = $1
	`, name)

	var metadataBytes []byte
	err = row.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.IsActive,
		&metadataBytes,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, core.TenantNotFoundError{Name: name}
		}
		return nil, err
	}

	// Parse metadata
	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &tenant.Metadata); err != nil {
			return nil, err
		}
	}

	// Get datasources
	rows, err := tx.Query(ctx, `
		SELECT id, dsn, role, pool_size, metadata, created_at, updated_at
		FROM datasources WHERE tenant_id = $1
		ORDER BY created_at
	`, tenant.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ds := core.Datasource{TenantID: tenant.ID}
		var metadataBytes []byte

		if err := rows.Scan(
			&ds.ID,
			&ds.DSN,
			&ds.Role,
			&ds.PoolSize,
			&metadataBytes,
			&ds.CreatedAt,
			&ds.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &ds.Metadata); err != nil {
				return nil, err
			}
		}

		tenant.Datasources = append(tenant.Datasources, ds)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return tenant, nil
}

// List retrieves all tenants with optional filtering
func (r *TenantRepository) List(ctx context.Context) ([]core.Tenant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, is_active, metadata, created_at, updated_at 
		FROM tenants 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []core.Tenant
	for rows.Next() {
		tenant := core.Tenant{}
		var metadataBytes []byte

		if err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.IsActive,
			&metadataBytes,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Parse metadata
		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &tenant.Metadata); err != nil {
				return nil, err
			}
		}

		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load datasources for each tenant
	for i := range tenants {
		datasources, err := r.getDatasourcesByTenantID(ctx, tenants[i].ID)
		if err != nil {
			return nil, err
		}
		tenants[i].Datasources = datasources
	}

	return tenants, nil
}

// Create creates a new tenant with its datasources
func (r *TenantRepository) Create(ctx context.Context, tenant *core.Tenant) error {
	// Validate tenant before creating
	if err := tenant.Validate(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Serialize metadata
	var metadataBytes []byte
	if tenant.Metadata != nil {
		metadataBytes, err = json.Marshal(tenant.Metadata)
		if err != nil {
			return err
		}
	}

	// Set timestamps
	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	// Insert tenant
	_, err = tx.Exec(ctx, `
		INSERT INTO tenants (id, name, is_active, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tenant.ID, tenant.Name, tenant.IsActive, metadataBytes, tenant.CreatedAt, tenant.UpdatedAt)
	if err != nil {
		return mapPostgreSQLError(err)
	}

	// Insert datasources
	for i := range tenant.Datasources {
		ds := &tenant.Datasources[i]
		ds.TenantID = tenant.ID
		ds.CreatedAt = now
		ds.UpdatedAt = now

		if err := ds.Validate(); err != nil {
			return err
		}

		var dsMetadataBytes []byte
		if ds.Metadata != nil {
			dsMetadataBytes, err = json.Marshal(ds.Metadata)
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO datasources (id, tenant_id, dsn, role, pool_size, metadata, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, ds.ID, ds.TenantID, ds.DSN, ds.Role, ds.PoolSize, dsMetadataBytes, ds.CreatedAt, ds.UpdatedAt)
		if err != nil {
			return mapPostgreSQLError(err)
		}
	}

	return tx.Commit(ctx)
}

// Update updates an existing tenant and its datasources
func (r *TenantRepository) Update(ctx context.Context, tenant *core.Tenant) error {
	// Validate tenant before updating
	if err := tenant.Validate(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check if tenant exists
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)", tenant.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return core.TenantNotFoundError{Name: tenant.Name}
	}

	// Serialize metadata
	var metadataBytes []byte
	if tenant.Metadata != nil {
		metadataBytes, err = json.Marshal(tenant.Metadata)
		if err != nil {
			return err
		}
	}

	// Update tenant
	tenant.UpdatedAt = time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE tenants 
		SET name = $2, is_active = $3, metadata = $4, updated_at = $5
		WHERE id = $1
	`, tenant.ID, tenant.Name, tenant.IsActive, metadataBytes, tenant.UpdatedAt)
	if err != nil {
		return mapPostgreSQLError(err)
	}

	// Delete existing datasources
	_, err = tx.Exec(ctx, "DELETE FROM datasources WHERE tenant_id = $1", tenant.ID)
	if err != nil {
		return err
	}

	// Insert new datasources
	for i := range tenant.Datasources {
		ds := &tenant.Datasources[i]
		ds.TenantID = tenant.ID
		ds.UpdatedAt = time.Now()
		if ds.CreatedAt.IsZero() {
			ds.CreatedAt = ds.UpdatedAt
		}

		if err := ds.Validate(); err != nil {
			return err
		}

		var dsMetadataBytes []byte
		if ds.Metadata != nil {
			dsMetadataBytes, err = json.Marshal(ds.Metadata)
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO datasources (id, tenant_id, dsn, role, pool_size, metadata, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, ds.ID, ds.TenantID, ds.DSN, ds.Role, ds.PoolSize, dsMetadataBytes, ds.CreatedAt, ds.UpdatedAt)
		if err != nil {
			return mapPostgreSQLError(err)
		}
	}

	return tx.Commit(ctx)
}

// Delete removes a tenant and all its datasources
func (r *TenantRepository) Delete(ctx context.Context, name string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check if tenant exists
	var tenantID string
	err = tx.QueryRow(ctx, "SELECT id FROM tenants WHERE name = $1", name).Scan(&tenantID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return core.TenantNotFoundError{Name: name}
		}
		return err
	}

	// Delete tenant (datasources will be deleted by CASCADE)
	result, err := tx.Exec(ctx, "DELETE FROM tenants WHERE name = $1", name)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return core.TenantNotFoundError{Name: name}
	}

	return tx.Commit(ctx)
}

// getDatasourcesByTenantID is a helper method to get datasources for a tenant
func (r *TenantRepository) getDatasourcesByTenantID(ctx context.Context, tenantID string) ([]core.Datasource, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, dsn, role, pool_size, metadata, created_at, updated_at
		FROM datasources WHERE tenant_id = $1
		ORDER BY created_at
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasources []core.Datasource
	for rows.Next() {
		ds := core.Datasource{TenantID: tenantID}
		var metadataBytes []byte

		if err := rows.Scan(
			&ds.ID,
			&ds.DSN,
			&ds.Role,
			&ds.PoolSize,
			&metadataBytes,
			&ds.CreatedAt,
			&ds.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &ds.Metadata); err != nil {
				return nil, err
			}
		}

		datasources = append(datasources, ds)
	}

	return datasources, rows.Err()
}
