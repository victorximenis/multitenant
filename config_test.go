package multitenant

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{}
	for _, key := range []string{
		"MULTITENANT_DATABASE_TYPE",
		"MULTITENANT_DATABASE_DSN",
		"MULTITENANT_REDIS_URL",
		"MULTITENANT_CACHE_TTL",
		"MULTITENANT_HEADER_NAME",
		"MULTITENANT_POOL_SIZE",
		"MULTITENANT_MAX_RETRIES",
		"MULTITENANT_RETRY_DELAY",
		"MULTITENANT_LOG_LEVEL",
	} {
		originalEnv[key] = os.Getenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Default values", func(t *testing.T) {
		// Clear all environment variables
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		// Set required values
		os.Setenv("MULTITENANT_DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("MULTITENANT_REDIS_URL", "redis://localhost:6379")

		config, err := LoadConfigFromEnv()
		assert.NoError(t, err)
		assert.Equal(t, PostgreSQL, config.DatabaseType)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", config.DatabaseDSN)
		assert.Equal(t, "redis://localhost:6379", config.RedisURL)
		assert.Equal(t, 5*time.Minute, config.CacheTTL)
		assert.Equal(t, "X-Tenant-Id", config.HeaderName)
		assert.Equal(t, 10, config.PoolSize)
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 1*time.Second, config.RetryDelay)
		assert.Equal(t, "info", config.LogLevel)
	})

	t.Run("Custom values", func(t *testing.T) {
		// Set custom values
		os.Setenv("MULTITENANT_DATABASE_TYPE", "mongodb")
		os.Setenv("MULTITENANT_DATABASE_DSN", "mongodb://localhost:27017/test")
		os.Setenv("MULTITENANT_REDIS_URL", "redis://localhost:6380")
		os.Setenv("MULTITENANT_CACHE_TTL", "10m")
		os.Setenv("MULTITENANT_HEADER_NAME", "X-Custom-Tenant")
		os.Setenv("MULTITENANT_POOL_SIZE", "20")
		os.Setenv("MULTITENANT_MAX_RETRIES", "5")
		os.Setenv("MULTITENANT_RETRY_DELAY", "2s")
		os.Setenv("MULTITENANT_LOG_LEVEL", "debug")

		config, err := LoadConfigFromEnv()
		assert.NoError(t, err)
		assert.Equal(t, MongoDB, config.DatabaseType)
		assert.Equal(t, "mongodb://localhost:27017/test", config.DatabaseDSN)
		assert.Equal(t, "redis://localhost:6380", config.RedisURL)
		assert.Equal(t, 10*time.Minute, config.CacheTTL)
		assert.Equal(t, "X-Custom-Tenant", config.HeaderName)
		assert.Equal(t, 20, config.PoolSize)
		assert.Equal(t, 5, config.MaxRetries)
		assert.Equal(t, 2*time.Second, config.RetryDelay)
		assert.Equal(t, "debug", config.LogLevel)
	})

	t.Run("Invalid database type", func(t *testing.T) {
		os.Setenv("MULTITENANT_DATABASE_TYPE", "invalid")
		os.Setenv("MULTITENANT_DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("MULTITENANT_REDIS_URL", "redis://localhost:6379")

		_, err := LoadConfigFromEnv()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid database type: invalid")
	})

	t.Run("Invalid cache TTL", func(t *testing.T) {
		// Clear all environment variables first
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("MULTITENANT_DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("MULTITENANT_REDIS_URL", "redis://localhost:6379")
		os.Setenv("MULTITENANT_CACHE_TTL", "invalid")

		_, err := LoadConfigFromEnv()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cache TTL: invalid")
	})

	t.Run("Invalid pool size", func(t *testing.T) {
		// Clear all environment variables first
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("MULTITENANT_DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("MULTITENANT_REDIS_URL", "redis://localhost:6379")
		os.Setenv("MULTITENANT_POOL_SIZE", "invalid")

		_, err := LoadConfigFromEnv()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pool size: invalid")
	})

	t.Run("Invalid retry delay", func(t *testing.T) {
		// Clear all environment variables first
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("MULTITENANT_DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("MULTITENANT_REDIS_URL", "redis://localhost:6379")
		os.Setenv("MULTITENANT_RETRY_DELAY", "invalid")

		_, err := LoadConfigFromEnv()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid retry delay: invalid")
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     5 * time.Minute,
			HeaderName:   "X-Tenant-Id",
			PoolSize:     10,
			MaxRetries:   3,
			RetryDelay:   1 * time.Second,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing database DSN", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     5 * time.Minute,
			PoolSize:     10,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database DSN is required")
	})

	t.Run("Missing Redis URL", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			CacheTTL:     5 * time.Minute,
			PoolSize:     10,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis URL is required")
	})

	t.Run("Invalid pool size", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     5 * time.Minute,
			PoolSize:     0,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool size must be greater than 0")
	})

	t.Run("Invalid cache TTL", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     0,
			PoolSize:     10,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache TTL must be greater than 0")
	})

	t.Run("Invalid log level", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     5 * time.Minute,
			PoolSize:     10,
			LogLevel:     "invalid",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log level: invalid")
	})

	t.Run("Negative max retries", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     5 * time.Minute,
			PoolSize:     10,
			MaxRetries:   -1,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max retries cannot be negative")
	})

	t.Run("Negative retry delay", func(t *testing.T) {
		config := &Config{
			DatabaseType: PostgreSQL,
			DatabaseDSN:  "postgres://user:pass@localhost:5432/db",
			RedisURL:     "redis://localhost:6379",
			CacheTTL:     5 * time.Minute,
			PoolSize:     10,
			RetryDelay:   -1 * time.Second,
			LogLevel:     "info",
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retry delay cannot be negative")
	})
}
