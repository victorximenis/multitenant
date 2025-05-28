package tenantcontext

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorximenis/multitenant/core"
)

func TestWithTimeout(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	timeoutCtx, cancel := WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Verify tenant is preserved
	retrieved, ok := GetTenant(timeoutCtx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)

	// Verify timeout works
	select {
	case <-timeoutCtx.Done():
		assert.Equal(t, context.DeadlineExceeded, timeoutCtx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout should have occurred")
	}
}

func TestWithTimeoutNoTenant(t *testing.T) {
	ctx := context.Background()

	timeoutCtx, cancel := WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Verify no tenant in context
	_, ok := GetTenant(timeoutCtx)
	assert.False(t, ok)
}

func TestWithCancel(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	cancelCtx, cancel := WithCancel(ctx)

	// Verify tenant is preserved
	retrieved, ok := GetTenant(cancelCtx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)

	// Test cancellation
	cancel()

	select {
	case <-cancelCtx.Done():
		assert.Equal(t, context.Canceled, cancelCtx.Err())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should have been cancelled")
	}
}

func TestWithDeadline(t *testing.T) {
	tenant := NewTestTenant("test-tenant")
	ctx := WithTenant(context.Background(), tenant)

	deadline := time.Now().Add(100 * time.Millisecond)
	deadlineCtx, cancel := WithDeadline(ctx, deadline)
	defer cancel()

	// Verify tenant is preserved
	retrieved, ok := GetTenant(deadlineCtx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)

	// Verify deadline works
	select {
	case <-deadlineCtx.Done():
		assert.Equal(t, context.DeadlineExceeded, deadlineCtx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Fatal("deadline should have been exceeded")
	}
}

func TestValidateTenantContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
		errType interface{}
	}{
		{
			name:    "valid active tenant",
			ctx:     NewTestContext(),
			wantErr: false,
		},
		{
			name:    "no tenant in context",
			ctx:     context.Background(),
			wantErr: true,
			errType: core.TenantNotFoundError{},
		},
		{
			name:    "inactive tenant",
			ctx:     NewTestContextWithInactiveTenant("inactive"),
			wantErr: true,
			errType: core.TenantInactiveError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTenantContext(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCopyTenantToContext(t *testing.T) {
	tenant := NewTestTenant("source-tenant")
	sourceCtx := WithTenant(context.Background(), tenant)
	targetCtx := context.Background()

	// Copy tenant from source to target
	resultCtx := CopyTenantToContext(sourceCtx, targetCtx)

	// Verify tenant was copied
	retrieved, ok := GetTenant(resultCtx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)

	// Test with no tenant in source
	emptyCtx := context.Background()
	resultCtx = CopyTenantToContext(emptyCtx, targetCtx)

	// Should return target context unchanged
	_, ok = GetTenant(resultCtx)
	assert.False(t, ok)
}

func TestExecuteWithTenant(t *testing.T) {
	tenant := NewTestTenant("execution-tenant")
	ctx := context.Background()

	var capturedTenantName string
	err := ExecuteWithTenant(ctx, tenant, func(execCtx context.Context) error {
		capturedTenantName = GetCurrentTenantName(execCtx)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "execution-tenant", capturedTenantName)

	// Verify original context is unchanged
	_, ok := GetTenant(ctx)
	assert.False(t, ok)
}

func TestExecuteWithTenantFunc(t *testing.T) {
	tenant := NewTestTenant("execution-tenant")
	ctx := context.Background()

	result, err := ExecuteWithTenantFunc(ctx, tenant, func(execCtx context.Context) (string, error) {
		return GetCurrentTenantName(execCtx), nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "execution-tenant", result)
}

func TestExecuteWithTenantError(t *testing.T) {
	tenant := NewTestTenant("execution-tenant")
	ctx := context.Background()

	expectedErr := assert.AnError
	err := ExecuteWithTenant(ctx, tenant, func(execCtx context.Context) error {
		return expectedErr
	})

	assert.Equal(t, expectedErr, err)
}

func TestContextChaining(t *testing.T) {
	tenant := NewTestTenant("chain-tenant")
	ctx := WithTenant(context.Background(), tenant)

	// Chain multiple context operations
	timeoutCtx, cancel1 := WithTimeout(ctx, 1*time.Second)
	defer cancel1()

	cancelCtx, cancel2 := WithCancel(timeoutCtx)
	defer cancel2()

	// Verify tenant is preserved through the chain
	retrieved, ok := GetTenant(cancelCtx)
	assert.True(t, ok)
	assert.Equal(t, tenant, retrieved)
	assert.Equal(t, "chain-tenant", GetCurrentTenantName(cancelCtx))
}
