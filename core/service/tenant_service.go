package service

import (
	"context"
	"time"

	"github.com/victorximenis/multitenant/core"
)

type TenantService struct {
	repo  core.TenantRepository
	cache core.TenantCache
	ttl   time.Duration
}

type Config struct {
	Repository core.TenantRepository
	Cache      core.TenantCache
	CacheTTL   time.Duration
}

func NewTenantService(config Config) *TenantService {
	ttl := config.CacheTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	return &TenantService{
		repo:  config.Repository,
		cache: config.Cache,
		ttl:   ttl,
	}
}

func (s *TenantService) GetTenant(ctx context.Context, name string) (*core.Tenant, error) {
	// Try to get from cache first
	tenant, err := s.cache.Get(ctx, name)
	if err == nil {
		return tenant, nil
	}

	// If not in cache or error, try repository
	tenant, err = s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Cache the tenant for future requests
	if err := s.cache.Set(ctx, tenant, s.ttl); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return tenant, nil
}

func (s *TenantService) ListTenants(ctx context.Context) ([]core.Tenant, error) {
	// List always goes directly to repository
	return s.repo.List(ctx)
}

func (s *TenantService) CreateTenant(ctx context.Context, tenant *core.Tenant) error {
	if err := s.repo.Create(ctx, tenant); err != nil {
		return err
	}

	// Cache the new tenant
	return s.cache.Set(ctx, tenant, s.ttl)
}

func (s *TenantService) UpdateTenant(ctx context.Context, tenant *core.Tenant) error {
	if err := s.repo.Update(ctx, tenant); err != nil {
		return err
	}

	// Update cache
	return s.cache.Set(ctx, tenant, s.ttl)
}

func (s *TenantService) DeleteTenant(ctx context.Context, id string) error {
	// Get tenant first to know the name for cache invalidation
	tenants, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	var tenantName string
	for _, t := range tenants {
		if t.ID == id {
			tenantName = t.Name
			break
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache if we found the tenant name
	if tenantName != "" {
		return s.cache.Delete(ctx, tenantName)
	}

	return nil
}
