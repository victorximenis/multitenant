package cli

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

func TestWorker_SingleTenant(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})

	// Test single tenant worker
	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    false,
		TenantName:    "test-tenant",
		PollInterval:  10 * time.Millisecond,
	})

	ctx := context.Background()
	var processed bool
	var processedTenant string
	var mu sync.Mutex

	err := worker.Start(ctx, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		tenant, ok := tenantcontext.GetTenant(ctx)
		assert.True(t, ok)
		processedTenant = tenant.Name
		processed = true
		return nil
	})

	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)
	worker.Shutdown()

	mu.Lock()
	assert.True(t, processed)
	assert.Equal(t, "test-tenant", processedTenant)
	mu.Unlock()
}

func TestWorker_AllTenants(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id-2",
		Name:     "test-tenant-2",
		IsActive: true,
	})
	mockService.AddTenant(&core.Tenant{
		ID:       "inactive-id",
		Name:     "inactive-tenant",
		IsActive: false,
	})

	// Test all tenants worker
	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    true,
		PollInterval:  10 * time.Millisecond,
	})

	ctx := context.Background()
	processedTenants := make(map[string]bool)
	var mu sync.Mutex

	err := worker.Start(ctx, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		tenant, _ := tenantcontext.GetTenant(ctx)
		processedTenants[tenant.Name] = true
		return nil
	})

	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)
	worker.Shutdown()

	mu.Lock()
	assert.True(t, processedTenants["test-tenant"])
	assert.True(t, processedTenants["test-tenant-2"])
	assert.False(t, processedTenants["inactive-tenant"]) // Should not process inactive tenants
	mu.Unlock()
}

func TestWorker_WithEnvironmentVariable(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "env-tenant",
		IsActive: true,
	})

	// Set environment variable
	os.Setenv("TEST_TENANT_ENV", "env-tenant")
	defer os.Unsetenv("TEST_TENANT_ENV")

	// Test worker with environment variable
	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    false,
		EnvVarName:    "TEST_TENANT_ENV",
		PollInterval:  10 * time.Millisecond,
	})

	ctx := context.Background()
	var processed bool
	var processedTenant string
	var mu sync.Mutex

	err := worker.Start(ctx, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		tenant, ok := tenantcontext.GetTenant(ctx)
		assert.True(t, ok)
		processedTenant = tenant.Name
		processed = true
		return nil
	})

	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)
	worker.Shutdown()

	mu.Lock()
	assert.True(t, processed)
	assert.Equal(t, "env-tenant", processedTenant)
	mu.Unlock()
}

func TestWorker_NoTenantSpecified(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()

	// Test worker without tenant specified
	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    false,
		PollInterval:  10 * time.Millisecond,
	})

	ctx := context.Background()
	var processed bool
	var mu sync.Mutex

	err := worker.Start(ctx, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		processed = true
		return nil
	})

	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)
	worker.Shutdown()

	mu.Lock()
	assert.False(t, processed) // Should not process anything
	mu.Unlock()
}

func TestWorker_ProcessingError(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})

	// Test worker with processing error
	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    false,
		TenantName:    "test-tenant",
		PollInterval:  10 * time.Millisecond,
	})

	ctx := context.Background()
	var processed bool
	var mu sync.Mutex

	err := worker.Start(ctx, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		processed = true
		return assert.AnError // Return an error to test error handling
	})

	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)
	worker.Shutdown()

	mu.Lock()
	assert.True(t, processed) // Should still process, but handle the error
	mu.Unlock()
}

func TestNewWorker(t *testing.T) {
	mockService := NewMockTenantService()

	// Test with default poll interval
	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
	})
	assert.Equal(t, 1*time.Minute, worker.pollInterval)

	// Test with custom poll interval
	worker = NewWorker(WorkerConfig{
		TenantService: mockService,
		PollInterval:  30 * time.Second,
	})
	assert.Equal(t, 30*time.Second, worker.pollInterval)

	// Test configuration
	worker = NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    true,
		TenantName:    "test-tenant",
		EnvVarName:    "CUSTOM_ENV",
		PollInterval:  5 * time.Second,
	})

	assert.True(t, worker.processAll)
	assert.Equal(t, "test-tenant", worker.tenantName)
	assert.Equal(t, 5*time.Second, worker.pollInterval)
	assert.Equal(t, "CUSTOM_ENV", worker.resolver.envVarName)
}

func TestWorker_Shutdown(t *testing.T) {
	mockService := NewMockTenantService()
	mockService.AddTenant(&core.Tenant{
		ID:       "test-id",
		Name:     "test-tenant",
		IsActive: true,
	})

	worker := NewWorker(WorkerConfig{
		TenantService: mockService,
		ProcessAll:    false,
		TenantName:    "test-tenant",
		PollInterval:  100 * time.Millisecond, // Longer interval to test shutdown
	})

	ctx := context.Background()
	processCount := 0
	var mu sync.Mutex

	err := worker.Start(ctx, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		processCount++
		return nil
	})

	assert.NoError(t, err)

	// Wait for first processing
	time.Sleep(10 * time.Millisecond)

	// Shutdown before next processing cycle
	worker.Shutdown()

	// Wait a bit more to ensure no additional processing
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	// Should have processed at least once (immediately on start)
	// but not continue processing after shutdown
	assert.GreaterOrEqual(t, processCount, 1)
	assert.LessOrEqual(t, processCount, 2) // At most 2 (start + one interval)
	mu.Unlock()
}
