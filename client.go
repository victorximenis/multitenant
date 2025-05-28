package multitenant

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/core/service"
	"github.com/victorximenis/multitenant/infra/connection"
	"github.com/victorximenis/multitenant/infra/mongodb"
	"github.com/victorximenis/multitenant/infra/postgres"
	"github.com/victorximenis/multitenant/infra/redis"
	"github.com/victorximenis/multitenant/interfaces/cli"
	httpMiddleware "github.com/victorximenis/multitenant/interfaces/http"
)

// MultitenantClient is the main client for the multitenant library
type MultitenantClient struct {
	config            *Config
	tenantService     core.TenantService
	connectionManager *connection.ConnectionManager
	tenantResolver    *cli.TenantResolver
}

// NewMultitenantClient creates a new multitenant client
func NewMultitenantClient(ctx context.Context, config *Config) (*MultitenantClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create repository based on database type
	var repository core.TenantRepository
	var err error

	if config.DatabaseType == PostgreSQL {
		repository, err = postgres.NewTenantRepository(ctx, config.DatabaseDSN)
	} else {
		repository, err = mongodb.NewTenantRepository(ctx, config.DatabaseDSN)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant repository: %w", err)
	}

	// Create cache
	cache, err := redis.NewTenantCache(ctx, redis.Config{
		RedisURL: config.RedisURL,
		TTL:      config.CacheTTL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant cache: %w", err)
	}

	// Create tenant service
	tenantService := service.NewTenantService(service.Config{
		Repository: repository,
		Cache:      cache,
		CacheTTL:   config.CacheTTL,
	})

	// Create connection manager
	connectionManager := connection.NewConnectionManager(tenantService, connection.ConnectionConfig{
		MaxPoolSize: config.PoolSize,
		MinPoolSize: 2,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 1 * time.Hour,
		HealthCheck: 1 * time.Minute,
	})

	// Create tenant resolver
	tenantResolver := cli.NewTenantResolver(tenantService, "TENANT_NAME")

	return &MultitenantClient{
		config:            config,
		tenantService:     tenantService,
		connectionManager: connectionManager,
		tenantResolver:    tenantResolver,
	}, nil
}

// GetTenantService returns the tenant service
func (c *MultitenantClient) GetTenantService() core.TenantService {
	return c.tenantService
}

// GetConnectionManager returns the connection manager
func (c *MultitenantClient) GetConnectionManager() *connection.ConnectionManager {
	return c.connectionManager
}

// GetTenantResolver returns the tenant resolver
func (c *MultitenantClient) GetTenantResolver() *cli.TenantResolver {
	return c.tenantResolver
}

// GinMiddleware returns a Gin middleware for tenant resolution
func (c *MultitenantClient) GinMiddleware() gin.HandlerFunc {
	return httpMiddleware.TenantMiddleware(httpMiddleware.GinMiddlewareConfig{
		TenantService: c.tenantService,
		HeaderName:    c.config.HeaderName,
	})
}

// FiberMiddleware returns a Fiber middleware for tenant resolution
func (c *MultitenantClient) FiberMiddleware() fiber.Handler {
	return httpMiddleware.FiberTenantMiddleware(httpMiddleware.FiberMiddlewareConfig{
		TenantService: c.tenantService,
		HeaderName:    c.config.HeaderName,
	})
}

// ChiMiddleware returns a Chi middleware for tenant resolution
func (c *MultitenantClient) ChiMiddleware() func(http.Handler) http.Handler {
	return httpMiddleware.ChiTenantMiddleware(httpMiddleware.ChiMiddlewareConfig{
		TenantService: c.tenantService,
		HeaderName:    c.config.HeaderName,
	})
}

// Close closes all connections and resources
func (c *MultitenantClient) Close(ctx context.Context) error {
	c.connectionManager.CloseAll(ctx)
	return nil
}
