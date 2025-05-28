package core

import (
	"context"
	"time"
)

// TenantRepository defines the interface for tenant data persistence operations
type TenantRepository interface {
	GetByName(ctx context.Context, name string) (*Tenant, error)
	List(ctx context.Context) ([]Tenant, error)
	Create(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
}

// TenantCache defines the interface for tenant caching operations
type TenantCache interface {
	Get(ctx context.Context, name string) (*Tenant, error)
	Set(ctx context.Context, tenant *Tenant, ttl time.Duration) error
	Delete(ctx context.Context, name string) error
}

// TenantService defines the interface for tenant business logic operations
type TenantService interface {
	GetTenant(ctx context.Context, name string) (*Tenant, error)
	ListTenants(ctx context.Context) ([]Tenant, error)
	CreateTenant(ctx context.Context, tenant *Tenant) error
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, id string) error
}
