package http

import (
	"context"

	"github.com/victorximenis/multitenant/core"
)

// MockTenantService is a mock implementation of core.TenantService for testing
type MockTenantService struct {
	tenants map[string]*core.Tenant
}

// NewMockTenantService creates a new mock tenant service with predefined tenants
func NewMockTenantService() *MockTenantService {
	return &MockTenantService{
		tenants: map[string]*core.Tenant{
			"test-tenant": {
				ID:       "test-id",
				Name:     "test-tenant",
				IsActive: true,
			},
			"inactive-tenant": {
				ID:       "inactive-id",
				Name:     "inactive-tenant",
				IsActive: false,
			},
		},
	}
}

// GetTenant implements core.TenantService
func (m *MockTenantService) GetTenant(ctx context.Context, name string) (*core.Tenant, error) {
	tenant, exists := m.tenants[name]
	if !exists {
		return nil, core.TenantNotFoundError{Name: name}
	}

	if !tenant.IsActive {
		return nil, core.TenantInactiveError{Name: name}
	}

	return tenant, nil
}

// ListTenants implements core.TenantService
func (m *MockTenantService) ListTenants(ctx context.Context) ([]core.Tenant, error) {
	var tenants []core.Tenant
	for _, tenant := range m.tenants {
		tenants = append(tenants, *tenant)
	}
	return tenants, nil
}

// CreateTenant implements core.TenantService
func (m *MockTenantService) CreateTenant(ctx context.Context, tenant *core.Tenant) error {
	m.tenants[tenant.Name] = tenant
	return nil
}

// UpdateTenant implements core.TenantService
func (m *MockTenantService) UpdateTenant(ctx context.Context, tenant *core.Tenant) error {
	m.tenants[tenant.Name] = tenant
	return nil
}

// DeleteTenant implements core.TenantService
func (m *MockTenantService) DeleteTenant(ctx context.Context, id string) error {
	for name, tenant := range m.tenants {
		if tenant.ID == id {
			delete(m.tenants, name)
			return nil
		}
	}
	return core.TenantNotFoundError{Name: id}
}
