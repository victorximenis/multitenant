package multitenant

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/core/service"
	"github.com/victorximenis/multitenant/infra/connection"
	redisCache "github.com/victorximenis/multitenant/infra/redis"
	"github.com/victorximenis/multitenant/tenantcontext"
)

// TestIntegrationWithExistingContainers testa a integração usando containers existentes
func TestIntegrationWithExistingContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test Redis connection (usando container existente na porta 6379)
	t.Run("Redis Connection", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		defer rdb.Close()

		// Test ping
		pong, err := rdb.Ping(ctx).Result()
		if err != nil {
			t.Skipf("Redis not available: %v", err)
		}
		assert.Equal(t, "PONG", pong)
	})

	// Test PostgreSQL connection (usando container existente na porta 5432)
	t.Run("PostgreSQL Connection", func(t *testing.T) {
		// Usando credenciais corretas do container postgres
		dsn := "postgres://dev_user:dev_password@localhost:5432/multitenant_db?sslmode=disable"

		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Skipf("PostgreSQL not available: %v", err)
		}
		defer pool.Close()

		// Test ping
		err = pool.Ping(ctx)
		assert.NoError(t, err)
	})

	// Test TenantService integration
	t.Run("TenantService Integration", func(t *testing.T) {
		// Create mock implementations for testing
		mockRepo := &mockTenantRepository{}

		// Create Redis cache
		cache, err := redisCache.NewTenantCache(ctx, redisCache.Config{
			RedisURL: "redis://localhost:6379",
			TTL:      5 * time.Minute,
		})
		if err != nil {
			t.Skipf("Redis cache not available: %v", err)
		}

		// Create TenantService
		tenantService := service.NewTenantService(service.Config{
			Repository: mockRepo,
			Cache:      cache,
			CacheTTL:   5 * time.Minute,
		})

		// Test tenant creation and retrieval
		tenant := &core.Tenant{
			ID:   "test-tenant-id",
			Name: "test-tenant",
		}

		err = tenantService.CreateTenant(ctx, tenant)
		assert.NoError(t, err)

		// Retrieve tenant
		retrieved, err := tenantService.GetTenant(ctx, "test-tenant")
		assert.NoError(t, err)
		assert.Equal(t, tenant.ID, retrieved.ID)
		assert.Equal(t, tenant.Name, retrieved.Name)
	})

	// Test ConnectionManager
	t.Run("ConnectionManager Integration", func(t *testing.T) {
		// Create mock tenant service
		mockTenantService := &mockTenantService{}

		// Create connection manager
		connManager := connection.NewConnectionManager(
			mockTenantService,
			connection.DefaultConnectionConfig(),
		)

		// Test with tenant context
		tenant := &core.Tenant{
			ID:   "test-tenant",
			Name: "test-tenant",
			Datasources: []core.Datasource{
				{
					ID:       "ds1",
					TenantID: "test-tenant",
					DSN:      "postgres://dev_user:dev_password@localhost:5432/multitenant_db?sslmode=disable",
					Role:     "rw",
					PoolSize: 5,
				},
			},
		}

		ctx = tenantcontext.WithTenant(ctx, tenant)

		// Test PostgreSQL pool retrieval
		pool, err := connManager.GetPostgresPool(ctx, "rw")
		if err != nil {
			t.Logf("PostgreSQL pool test skipped: %v", err)
		} else {
			assert.NotNil(t, pool)
			// Test ping
			err = pool.Ping(ctx)
			assert.NoError(t, err)
		}
	})
}

// Mock implementations for testing
type mockTenantRepository struct{}

func (m *mockTenantRepository) Create(ctx context.Context, tenant *core.Tenant) error {
	return nil
}

func (m *mockTenantRepository) GetByName(ctx context.Context, name string) (*core.Tenant, error) {
	return &core.Tenant{
		ID:   "test-tenant-id",
		Name: name,
	}, nil
}

func (m *mockTenantRepository) List(ctx context.Context) ([]core.Tenant, error) {
	return []core.Tenant{}, nil
}

func (m *mockTenantRepository) Update(ctx context.Context, tenant *core.Tenant) error {
	return nil
}

func (m *mockTenantRepository) Delete(ctx context.Context, id string) error {
	return nil
}

type mockTenantService struct{}

func (m *mockTenantService) GetTenant(ctx context.Context, name string) (*core.Tenant, error) {
	return &core.Tenant{
		ID:   "test-tenant",
		Name: name,
		Datasources: []core.Datasource{
			{
				ID:       "ds1",
				TenantID: "test-tenant",
				DSN:      "postgres://dev_user:dev_password@localhost:5432/multitenant_db?sslmode=disable",
				Role:     "rw",
				PoolSize: 5,
			},
		},
	}, nil
}

func (m *mockTenantService) CreateTenant(ctx context.Context, tenant *core.Tenant) error {
	return nil
}

func (m *mockTenantService) UpdateTenant(ctx context.Context, tenant *core.Tenant) error {
	return nil
}

func (m *mockTenantService) DeleteTenant(ctx context.Context, id string) error {
	return nil
}

func (m *mockTenantService) ListTenants(ctx context.Context) ([]core.Tenant, error) {
	return []core.Tenant{}, nil
}
