package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenantNotFoundError(t *testing.T) {
	tenantName := "test-tenant"
	err := TenantNotFoundError{Name: tenantName}

	expectedMsg := "tenant not found: test-tenant"
	assert.Equal(t, expectedMsg, err.Error())

	// Test that it implements the error interface
	var _ error = err
}

func TestTenantInactiveError(t *testing.T) {
	tenantName := "inactive-tenant"
	err := TenantInactiveError{Name: tenantName}

	expectedMsg := "tenant is inactive: inactive-tenant"
	assert.Equal(t, expectedMsg, err.Error())

	// Test that it implements the error interface
	var _ error = err
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "TenantNotFoundError",
			err:         TenantNotFoundError{Name: "missing"},
			expectedMsg: "tenant not found: missing",
		},
		{
			name:        "TenantInactiveError",
			err:         TenantInactiveError{Name: "disabled"},
			expectedMsg: "tenant is inactive: disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
		})
	}
}
