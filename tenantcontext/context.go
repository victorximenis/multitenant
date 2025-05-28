package tenantcontext

import (
	"context"

	"github.com/victorximenis/multitenant/core"
)

// contextKey is a private type used for context keys to avoid collisions
type contextKey string

// tenantContextKey is the key used to store tenant information in context
const tenantContextKey contextKey = "tenant"

// WithTenant stores a tenant in the context
func WithTenant(ctx context.Context, tenant *core.Tenant) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, tenantContextKey, tenant)
}

// GetTenant retrieves the tenant from the context
func GetTenant(ctx context.Context) (*core.Tenant, bool) {
	if ctx == nil {
		return nil, false
	}

	tenant, ok := ctx.Value(tenantContextKey).(*core.Tenant)
	return tenant, ok
}

// MustGetTenant retrieves the tenant from the context or panics if not found
func MustGetTenant(ctx context.Context) *core.Tenant {
	tenant, ok := GetTenant(ctx)
	if !ok {
		panic("tenant not found in context")
	}
	return tenant
}

// GetCurrentTenantName returns the name of the current tenant or empty string
func GetCurrentTenantName(ctx context.Context) string {
	tenant, ok := GetTenant(ctx)
	if !ok {
		return ""
	}
	return tenant.Name
}

// GetCurrentTenantID returns the ID of the current tenant or empty string
func GetCurrentTenantID(ctx context.Context) string {
	tenant, ok := GetTenant(ctx)
	if !ok {
		return ""
	}
	return tenant.ID
}

// HasTenant checks if a tenant is present in the context
func HasTenant(ctx context.Context) bool {
	_, ok := GetTenant(ctx)
	return ok
}
