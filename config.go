package multitenant

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/victorximenis/multitenant/core"
)

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	MongoDB    DatabaseType = "mongodb"
)

type Config struct {
	// Database configuration
	DatabaseType DatabaseType `json:"database_type"`
	DatabaseDSN  string       `json:"database_dsn"`

	// Redis configuration
	RedisURL string        `json:"redis_url"`
	CacheTTL time.Duration `json:"cache_ttl"`

	// HTTP configuration
	HeaderName string `json:"header_name"`

	// Connection pool configuration
	PoolSize   int           `json:"pool_size"`
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`

	// Logging configuration
	LogLevel string `json:"log_level"`
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		DatabaseType: PostgreSQL,
		HeaderName:   "X-Tenant-Id",
		CacheTTL:     5 * time.Minute,
		PoolSize:     10,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
		LogLevel:     "info",
	}

	// Database configuration
	if dbType := os.Getenv("MULTITENANT_DATABASE_TYPE"); dbType != "" {
		if dbType == "postgres" {
			config.DatabaseType = PostgreSQL
		} else if dbType == "mongodb" {
			config.DatabaseType = MongoDB
		} else {
			return nil, core.ErrConfigInvalid("MULTITENANT_DATABASE_TYPE",
				fmt.Sprintf("invalid database type: %s (must be 'postgres' or 'mongodb')", dbType))
		}
	}

	if dbDSN := os.Getenv("MULTITENANT_DATABASE_DSN"); dbDSN != "" {
		config.DatabaseDSN = dbDSN
	}

	// Redis configuration
	if redisURL := os.Getenv("MULTITENANT_REDIS_URL"); redisURL != "" {
		config.RedisURL = redisURL
	}

	if cacheTTL := os.Getenv("MULTITENANT_CACHE_TTL"); cacheTTL != "" {
		ttl, err := time.ParseDuration(cacheTTL)
		if err != nil {
			return nil, core.ErrConfigInvalid("MULTITENANT_CACHE_TTL",
				fmt.Sprintf("invalid cache TTL format: %s (example: '5m', '1h')", cacheTTL)).
				WithCause(err)
		}
		config.CacheTTL = ttl
	}

	// HTTP configuration
	if headerName := os.Getenv("MULTITENANT_HEADER_NAME"); headerName != "" {
		config.HeaderName = headerName
	}

	// Connection pool configuration
	if poolSize := os.Getenv("MULTITENANT_POOL_SIZE"); poolSize != "" {
		size, err := strconv.Atoi(poolSize)
		if err != nil {
			return nil, core.ErrConfigInvalid("MULTITENANT_POOL_SIZE",
				fmt.Sprintf("invalid pool size: %s (must be a positive integer)", poolSize)).
				WithCause(err)
		}
		config.PoolSize = size
	}

	if maxRetries := os.Getenv("MULTITENANT_MAX_RETRIES"); maxRetries != "" {
		retries, err := strconv.Atoi(maxRetries)
		if err != nil {
			return nil, core.ErrConfigInvalid("MULTITENANT_MAX_RETRIES",
				fmt.Sprintf("invalid max retries: %s (must be a non-negative integer)", maxRetries)).
				WithCause(err)
		}
		config.MaxRetries = retries
	}

	if retryDelay := os.Getenv("MULTITENANT_RETRY_DELAY"); retryDelay != "" {
		delay, err := time.ParseDuration(retryDelay)
		if err != nil {
			return nil, core.ErrConfigInvalid("MULTITENANT_RETRY_DELAY",
				fmt.Sprintf("invalid retry delay format: %s (example: '1s', '500ms')", retryDelay)).
				WithCause(err)
		}
		config.RetryDelay = delay
	}

	// Logging configuration
	if logLevel := os.Getenv("MULTITENANT_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	return config, config.Validate()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Database DSN validation
	if c.DatabaseDSN == "" {
		return core.ErrConfigInvalid("DatabaseDSN", "database DSN is required")
	}

	// Validate DSN format based on database type
	if err := c.validateDSN(); err != nil {
		return err
	}

	// Redis URL validation
	if c.RedisURL == "" {
		return core.ErrConfigInvalid("RedisURL", "Redis URL is required")
	}

	if err := c.validateRedisURL(); err != nil {
		return err
	}

	// Pool size validation
	if c.PoolSize <= 0 {
		return core.ErrConfigInvalid("PoolSize",
			fmt.Sprintf("pool size must be greater than 0, got: %d", c.PoolSize))
	}

	if c.PoolSize > 100 {
		return core.ErrConfigInvalid("PoolSize",
			fmt.Sprintf("pool size too large (max 100), got: %d", c.PoolSize))
	}

	// Max retries validation
	if c.MaxRetries < 0 {
		return core.ErrConfigInvalid("MaxRetries",
			fmt.Sprintf("max retries cannot be negative, got: %d", c.MaxRetries))
	}

	if c.MaxRetries > 10 {
		return core.ErrConfigInvalid("MaxRetries",
			fmt.Sprintf("max retries too large (max 10), got: %d", c.MaxRetries))
	}

	// Cache TTL validation
	if c.CacheTTL <= 0 {
		return core.ErrConfigInvalid("CacheTTL",
			fmt.Sprintf("cache TTL must be greater than 0, got: %v", c.CacheTTL))
	}

	if c.CacheTTL > 24*time.Hour {
		return core.ErrConfigInvalid("CacheTTL",
			fmt.Sprintf("cache TTL too large (max 24h), got: %v", c.CacheTTL))
	}

	// Retry delay validation
	if c.RetryDelay < 0 {
		return core.ErrConfigInvalid("RetryDelay",
			fmt.Sprintf("retry delay cannot be negative, got: %v", c.RetryDelay))
	}

	if c.RetryDelay > 1*time.Minute {
		return core.ErrConfigInvalid("RetryDelay",
			fmt.Sprintf("retry delay too large (max 1m), got: %v", c.RetryDelay))
	}

	// Log level validation
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return core.ErrConfigInvalid("LogLevel",
			fmt.Sprintf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel))
	}

	// Header name validation
	if c.HeaderName == "" {
		return core.ErrConfigInvalid("HeaderName", "header name cannot be empty")
	}

	if len(c.HeaderName) > 100 {
		return core.ErrConfigInvalid("HeaderName",
			fmt.Sprintf("header name too long (max 100 chars), got: %d", len(c.HeaderName)))
	}

	return nil
}

// validateDSN validates the database DSN format
func (c *Config) validateDSN() error {
	switch c.DatabaseType {
	case PostgreSQL:
		return c.validatePostgresDSN()
	case MongoDB:
		return c.validateMongoDSN()
	default:
		return core.ErrConfigInvalid("DatabaseType",
			fmt.Sprintf("unsupported database type: %s", c.DatabaseType))
	}
}

// validatePostgresDSN validates PostgreSQL DSN format
func (c *Config) validatePostgresDSN() error {
	// Basic format check for PostgreSQL DSN
	if !strings.HasPrefix(c.DatabaseDSN, "postgres://") && !strings.HasPrefix(c.DatabaseDSN, "postgresql://") {
		return core.ErrConfigInvalid("DatabaseDSN",
			"PostgreSQL DSN must start with 'postgres://' or 'postgresql://'")
	}

	// Try to parse as URL to validate format
	_, err := url.Parse(c.DatabaseDSN)
	if err != nil {
		return core.ErrConfigInvalid("DatabaseDSN",
			fmt.Sprintf("invalid PostgreSQL DSN format: %v", err)).WithCause(err)
	}

	return nil
}

// validateMongoDSN validates MongoDB DSN format
func (c *Config) validateMongoDSN() error {
	// Basic format check for MongoDB DSN
	if !strings.HasPrefix(c.DatabaseDSN, "mongodb://") && !strings.HasPrefix(c.DatabaseDSN, "mongodb+srv://") {
		return core.ErrConfigInvalid("DatabaseDSN",
			"MongoDB DSN must start with 'mongodb://' or 'mongodb+srv://'")
	}

	// Try to parse as URL to validate format
	_, err := url.Parse(c.DatabaseDSN)
	if err != nil {
		return core.ErrConfigInvalid("DatabaseDSN",
			fmt.Sprintf("invalid MongoDB DSN format: %v", err)).WithCause(err)
	}

	return nil
}

// validateRedisURL validates Redis URL format
func (c *Config) validateRedisURL() error {
	// Basic format check for Redis URL
	if !strings.HasPrefix(c.RedisURL, "redis://") && !strings.HasPrefix(c.RedisURL, "rediss://") {
		return core.ErrConfigInvalid("RedisURL",
			"Redis URL must start with 'redis://' or 'rediss://'")
	}

	// Try to parse as URL to validate format
	_, err := url.Parse(c.RedisURL)
	if err != nil {
		return core.ErrConfigInvalid("RedisURL",
			fmt.Sprintf("invalid Redis URL format: %v", err)).WithCause(err)
	}

	return nil
}

// GetDatabaseHost extracts the host from the database DSN
func (c *Config) GetDatabaseHost() (string, error) {
	u, err := url.Parse(c.DatabaseDSN)
	if err != nil {
		return "", core.ErrConfigInvalid("DatabaseDSN", "failed to parse DSN").WithCause(err)
	}
	return u.Host, nil
}

// GetRedisHost extracts the host from the Redis URL
func (c *Config) GetRedisHost() (string, error) {
	u, err := url.Parse(c.RedisURL)
	if err != nil {
		return "", core.ErrConfigInvalid("RedisURL", "failed to parse Redis URL").WithCause(err)
	}
	return u.Host, nil
}

// IsDevelopment checks if the configuration is for development
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.LogLevel) == "debug"
}

// IsProduction checks if the configuration is for production
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.LogLevel) == "error" || strings.ToLower(c.LogLevel) == "warn"
}
