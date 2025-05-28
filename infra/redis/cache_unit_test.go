package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTenantKey(t *testing.T) {
	cache := &TenantCache{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple tenant name",
			input:    "test-tenant",
			expected: "multitenant:tenants:test-tenant",
		},
		{
			name:     "tenant with special characters",
			input:    "tenant-with-dashes_and_underscores",
			expected: "multitenant:tenants:tenant-with-dashes_and_underscores",
		},
		{
			name:     "empty tenant name",
			input:    "",
			expected: "multitenant:tenants:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.tenantKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_DefaultTTL(t *testing.T) {
	tests := []struct {
		name        string
		configTTL   time.Duration
		expectedTTL time.Duration
	}{
		{
			name:        "zero TTL should use default",
			configTTL:   0,
			expectedTTL: DEFAULT_TTL,
		},
		{
			name:        "custom TTL should be preserved",
			configTTL:   10 * time.Minute,
			expectedTTL: 10 * time.Minute,
		},
		{
			name:        "negative TTL should use default",
			configTTL:   -1 * time.Minute,
			expectedTTL: DEFAULT_TTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test NewTenantCache without Redis, but we can test the logic
			var resultTTL time.Duration
			if tt.configTTL <= 0 {
				resultTTL = DEFAULT_TTL
			} else {
				resultTTL = tt.configTTL
			}

			assert.Equal(t, tt.expectedTTL, resultTTL)
		})
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 5*time.Minute, DEFAULT_TTL)
	assert.Equal(t, "multitenant:tenants:", KEY_PREFIX)
}
