package tenantcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTracing(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	var executedTenantName string
	err := WithTracing(ctx, func(execCtx context.Context) error {
		executedTenantName = GetCurrentTenantName(execCtx)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "test-tenant", executedTenantName)
}

func TestWithTracingFunc(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	result, err := WithTracingFunc(ctx, func(execCtx context.Context) (string, error) {
		return GetCurrentTenantName(execCtx), nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "test-tenant", result)
}

func TestWithTracingError(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	expectedErr := assert.AnError
	err := WithTracing(ctx, func(execCtx context.Context) error {
		return expectedErr
	})

	assert.Equal(t, expectedErr, err)
}

func TestWithTracingFuncError(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	expectedErr := assert.AnError
	result, err := WithTracingFunc(ctx, func(execCtx context.Context) (string, error) {
		return "", expectedErr
	})

	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result)
}

func TestPropagateToSpanNoSpan(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	// Test with context that has no span - should not panic
	assert.NotPanics(t, func() {
		PropagateToSpan(ctx)
	})
}

func TestPropagateToSpanNoTenant(t *testing.T) {
	ctx := context.Background()

	// Test with context that has no tenant - should not panic
	assert.NotPanics(t, func() {
		PropagateToSpan(ctx)
	})
}

func TestTracingIntegration(t *testing.T) {
	// Test that tracing functions work correctly with tenant context
	tenant := CreateTestTenantWithDatasources("integration-tenant", 2)
	ctx := WithTenant(context.Background(), tenant)

	// Test nested execution with tracing
	var innerTenantName string
	err := WithTracing(ctx, func(outerCtx context.Context) error {
		return WithTracing(outerCtx, func(innerCtx context.Context) error {
			innerTenantName = GetCurrentTenantName(innerCtx)
			return nil
		})
	})

	assert.NoError(t, err)
	assert.Equal(t, "integration-tenant", innerTenantName)
}

func TestTracingWithGenericFunction(t *testing.T) {
	tenant := NewTestTenant("generic-tenant")
	ctx := WithTenant(context.Background(), tenant)

	// Test with different return types
	stringResult, err := WithTracingFunc(ctx, func(execCtx context.Context) (string, error) {
		return GetCurrentTenantName(execCtx), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "generic-tenant", stringResult)

	intResult, err := WithTracingFunc(ctx, func(execCtx context.Context) (int, error) {
		tenant, _ := GetTenant(execCtx)
		return len(tenant.Datasources), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, intResult)

	boolResult, err := WithTracingFunc(ctx, func(execCtx context.Context) (bool, error) {
		return HasTenant(execCtx), nil
	})
	assert.NoError(t, err)
	assert.True(t, boolResult)
}
