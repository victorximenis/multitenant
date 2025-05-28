package tenantcontext

import (
	"context"
	"time"

	"github.com/victorximenis/multitenant/core"
)

// WithTimeout creates a context with timeout while preserving tenant information
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	tenant, hasTenant := GetTenant(ctx)

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)

	if hasTenant {
		timeoutCtx = WithTenant(timeoutCtx, tenant)
	}

	return timeoutCtx, cancel
}

// WithCancel creates a cancellable context while preserving tenant information
func WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	tenant, hasTenant := GetTenant(ctx)

	cancelCtx, cancel := context.WithCancel(ctx)

	if hasTenant {
		cancelCtx = WithTenant(cancelCtx, tenant)
	}

	return cancelCtx, cancel
}

// WithDeadline creates a context with deadline while preserving tenant information
func WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	tenant, hasTenant := GetTenant(ctx)

	deadlineCtx, cancel := context.WithDeadline(ctx, deadline)

	if hasTenant {
		deadlineCtx = WithTenant(deadlineCtx, tenant)
	}

	return deadlineCtx, cancel
}

// ValidateTenantContext checks if the tenant in context is valid and active
func ValidateTenantContext(ctx context.Context) error {
	tenant, ok := GetTenant(ctx)
	if !ok {
		return core.TenantNotFoundError{Name: "context"}
	}

	if !tenant.IsActive {
		return core.TenantInactiveError{Name: tenant.Name}
	}

	return tenant.Validate()
}

// CopyTenantToContext copies tenant information from source context to target context
func CopyTenantToContext(source, target context.Context) context.Context {
	tenant, ok := GetTenant(source)
	if !ok {
		return target
	}

	return WithTenant(target, tenant)
}

// ExecuteWithTenant executes a function with a specific tenant in context
func ExecuteWithTenant(ctx context.Context, tenant *core.Tenant, fn func(context.Context) error) error {
	tenantCtx := WithTenant(ctx, tenant)
	return fn(tenantCtx)
}

// ExecuteWithTenantFunc executes a function with a specific tenant in context (generic version)
func ExecuteWithTenantFunc[T any](ctx context.Context, tenant *core.Tenant, fn func(context.Context) (T, error)) (T, error) {
	tenantCtx := WithTenant(ctx, tenant)
	return fn(tenantCtx)
}
