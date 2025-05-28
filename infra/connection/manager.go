package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	MongoDB    DatabaseType = "mongodb"
)

type ConnectionManager struct {
	tenantService core.TenantService
	pgPools       map[string]map[string]*pgxpool.Pool // tenant -> role -> pool
	mongoPools    map[string]map[string]*mongo.Client // tenant -> role -> client
	mu            sync.RWMutex
	defaultConfig ConnectionConfig
}

type ConnectionConfig struct {
	MaxPoolSize int
	MinPoolSize int
	MaxIdleTime time.Duration
	MaxLifetime time.Duration
	HealthCheck time.Duration
}

func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		MaxPoolSize: 10,
		MinPoolSize: 2,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 1 * time.Hour,
		HealthCheck: 1 * time.Minute,
	}
}

func NewConnectionManager(tenantService core.TenantService, config ConnectionConfig) *ConnectionManager {
	if config.MaxPoolSize == 0 {
		config = DefaultConnectionConfig()
	}

	return &ConnectionManager{
		tenantService: tenantService,
		pgPools:       make(map[string]map[string]*pgxpool.Pool),
		mongoPools:    make(map[string]map[string]*mongo.Client),
		defaultConfig: config,
	}
}

func (m *ConnectionManager) GetPostgresPool(ctx context.Context, role string) (*pgxpool.Pool, error) {
	tenant, ok := tenantcontext.GetTenant(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant not found in context")
	}

	return m.GetPostgresPoolForTenant(ctx, tenant.Name, role)
}

func (m *ConnectionManager) GetPostgresPoolForTenant(ctx context.Context, tenantName, role string) (*pgxpool.Pool, error) {
	// Check if we already have a pool for this tenant and role
	m.mu.RLock()
	if pools, ok := m.pgPools[tenantName]; ok {
		if pool, ok := pools[role]; ok {
			m.mu.RUnlock()
			return pool, nil
		}
	}
	m.mu.RUnlock()

	// Get tenant configuration
	tenant, err := m.tenantService.GetTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}

	// Find matching datasource
	var dsn string
	var poolSize int

	for _, ds := range tenant.Datasources {
		if ds.Role == role || ds.Role == "rw" {
			dsn = ds.DSN
			poolSize = ds.PoolSize
			break
		}
	}

	if dsn == "" {
		return nil, fmt.Errorf("no datasource found for tenant %s with role %s", tenantName, role)
	}

	// Create pool configuration
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if poolSize > 0 {
		poolConfig.MaxConns = int32(poolSize)
	} else {
		poolConfig.MaxConns = int32(m.defaultConfig.MaxPoolSize)
	}

	poolConfig.MinConns = int32(m.defaultConfig.MinPoolSize)
	poolConfig.MaxConnIdleTime = m.defaultConfig.MaxIdleTime
	poolConfig.MaxConnLifetime = m.defaultConfig.MaxLifetime
	poolConfig.HealthCheckPeriod = m.defaultConfig.HealthCheck

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	// Store pool for future use
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.pgPools[tenantName]; !ok {
		m.pgPools[tenantName] = make(map[string]*pgxpool.Pool)
	}

	m.pgPools[tenantName][role] = pool

	return pool, nil
}

func (m *ConnectionManager) GetMongoClient(ctx context.Context, role string) (*mongo.Client, error) {
	tenant, ok := tenantcontext.GetTenant(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant not found in context")
	}

	return m.GetMongoClientForTenant(ctx, tenant.Name, role)
}

func (m *ConnectionManager) GetMongoClientForTenant(ctx context.Context, tenantName, role string) (*mongo.Client, error) {
	// Check if we already have a client for this tenant and role
	m.mu.RLock()
	if clients, ok := m.mongoPools[tenantName]; ok {
		if client, ok := clients[role]; ok {
			m.mu.RUnlock()
			return client, nil
		}
	}
	m.mu.RUnlock()

	// Get tenant configuration
	tenant, err := m.tenantService.GetTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}

	// Find matching datasource
	var dsn string
	var poolSize int

	for _, ds := range tenant.Datasources {
		if ds.Role == role || ds.Role == "rw" {
			dsn = ds.DSN
			poolSize = ds.PoolSize
			break
		}
	}

	if dsn == "" {
		return nil, fmt.Errorf("no datasource found for tenant %s with role %s", tenantName, role)
	}

	// Create client options
	clientOptions := options.Client().ApplyURI(dsn)

	if poolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(poolSize))
	} else {
		clientOptions.SetMaxPoolSize(uint64(m.defaultConfig.MaxPoolSize))
	}

	clientOptions.SetMinPoolSize(uint64(m.defaultConfig.MinPoolSize))
	clientOptions.SetMaxConnIdleTime(m.defaultConfig.MaxIdleTime)

	// Create client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	// Store client for future use
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.mongoPools[tenantName]; !ok {
		m.mongoPools[tenantName] = make(map[string]*mongo.Client)
	}

	m.mongoPools[tenantName][role] = client

	return client, nil
}

func (m *ConnectionManager) CloseAll(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all PostgreSQL pools
	for tenant, pools := range m.pgPools {
		for role, pool := range pools {
			pool.Close()
			delete(pools, role)
		}
		delete(m.pgPools, tenant)
	}

	// Close all MongoDB clients
	for tenant, clients := range m.mongoPools {
		for role, client := range clients {
			client.Disconnect(ctx)
			delete(clients, role)
		}
		delete(m.mongoPools, tenant)
	}
}
