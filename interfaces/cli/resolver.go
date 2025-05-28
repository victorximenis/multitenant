package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

type TenantResolver struct {
	tenantService core.TenantService
	envVarName    string
}

func NewTenantResolver(tenantService core.TenantService, envVarName string) *TenantResolver {
	if envVarName == "" {
		envVarName = "TENANT_NAME"
	}

	return &TenantResolver{
		tenantService: tenantService,
		envVarName:    envVarName,
	}
}

// ResolveTenantFromEnv resolves a tenant from environment variables
func (r *TenantResolver) ResolveTenantFromEnv(ctx context.Context) (context.Context, error) {
	tenantName := os.Getenv(r.envVarName)
	if tenantName == "" {
		return ctx, fmt.Errorf("environment variable %s not set", r.envVarName)
	}

	return r.ResolveTenant(ctx, tenantName)
}

// ResolveTenant resolves a tenant by name
func (r *TenantResolver) ResolveTenant(ctx context.Context, tenantName string) (context.Context, error) {
	tenant, err := r.tenantService.GetTenant(ctx, tenantName)
	if err != nil {
		return ctx, err
	}

	return tenantcontext.WithTenant(ctx, tenant), nil
}

// ResolveTenantFromArgs resolves a tenant from command line arguments
func (r *TenantResolver) ResolveTenantFromArgs(ctx context.Context, args []string) (context.Context, error) {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--tenant=") {
			tenantName := strings.TrimPrefix(arg, "--tenant=")
			return r.ResolveTenant(ctx, tenantName)
		} else if arg == "--tenant" && i+1 < len(args) {
			return r.ResolveTenant(ctx, args[i+1])
		}
	}

	return ctx, fmt.Errorf("tenant argument not found (use --tenant=name or --tenant name)")
}

// WithTenant is a helper function to run a function with a resolved tenant context
func (r *TenantResolver) WithTenant(ctx context.Context, tenantName string, fn func(context.Context) error) error {
	ctx, err := r.ResolveTenant(ctx, tenantName)
	if err != nil {
		return err
	}

	return fn(ctx)
}

// ForEachTenant runs a function for each active tenant
func (r *TenantResolver) ForEachTenant(ctx context.Context, fn func(context.Context) error) error {
	tenants, err := r.tenantService.ListTenants(ctx)
	if err != nil {
		return err
	}

	for _, tenant := range tenants {
		if !tenant.IsActive {
			continue
		}

		tenantCtx := tenantcontext.WithTenant(ctx, &tenant)
		if err := fn(tenantCtx); err != nil {
			return fmt.Errorf("error processing tenant %s: %w", tenant.Name, err)
		}
	}

	return nil
}
