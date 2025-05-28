package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// createTablesSQL contains the DDL statements to create the required tables
const createTablesSQL = `
CREATE TABLE IF NOT EXISTS tenants (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  is_active BOOLEAN DEFAULT TRUE,
  metadata JSONB,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS datasources (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  dsn TEXT NOT NULL,
  role TEXT CHECK (role IN ('read', 'write', 'rw')) NOT NULL,
  pool_size INTEGER DEFAULT 10,
  metadata JSONB,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_active ON tenants(is_active);
CREATE INDEX IF NOT EXISTS idx_datasources_tenant_id ON datasources(tenant_id);
CREATE INDEX IF NOT EXISTS idx_datasources_role ON datasources(role);
`

// SetupSchema creates the required tables and indexes in the database
func SetupSchema(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, createTablesSQL)
	return err
}
