package cli

import (
	"context"

	"github.com/victorximenis/multitenant/core"
)

// MockTenantService is a mock implementation of core.TenantService for testing
type MockTenantService struct {
	tenants map[string]*core.Tenant
}

func NewMockTenantService() *MockTenantService {
	return &MockTenantService{
		tenants: make(map[string]*core.Tenant),
	}
}

func (m *MockTenantService) AddTenant(tenant *core.Tenant) {
	m.tenants[tenant.Name] = tenant
}

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

func (m *MockTenantService) ListTenants(ctx context.Context) ([]core.Tenant, error) {
	tenants := make([]core.Tenant, 0, len(m.tenants))
	for _, tenant := range m.tenants {
		tenants = append(tenants, *tenant)
	}
	return tenants, nil
}

func (m *MockTenantService) CreateTenant(ctx context.Context, tenant *core.Tenant) error {
	m.tenants[tenant.Name] = tenant
	return nil
}

func (m *MockTenantService) UpdateTenant(ctx context.Context, tenant *core.Tenant) error {
	if _, exists := m.tenants[tenant.Name]; !exists {
		return core.TenantNotFoundError{Name: tenant.Name}
	}
	m.tenants[tenant.Name] = tenant
	return nil
}

func (m *MockTenantService) DeleteTenant(ctx context.Context, id string) error {
	for name, tenant := range m.tenants {
		if tenant.ID == id {
			delete(m.tenants, name)
			return nil
		}
	}
	return core.TenantNotFoundError{Name: id}
}
