package multitenant

import (
	"fmt"
	"os"
	"strconv"
	"time"
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
			return nil, fmt.Errorf("invalid database type: %s", dbType)
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
			return nil, fmt.Errorf("invalid cache TTL: %s", cacheTTL)
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
			return nil, fmt.Errorf("invalid pool size: %s", poolSize)
		}
		config.PoolSize = size
	}

	if maxRetries := os.Getenv("MULTITENANT_MAX_RETRIES"); maxRetries != "" {
		retries, err := strconv.Atoi(maxRetries)
		if err != nil {
			return nil, fmt.Errorf("invalid max retries: %s", maxRetries)
		}
		config.MaxRetries = retries
	}

	if retryDelay := os.Getenv("MULTITENANT_RETRY_DELAY"); retryDelay != "" {
		delay, err := time.ParseDuration(retryDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid retry delay: %s", retryDelay)
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
	if c.DatabaseDSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	if c.RedisURL == "" {
		return fmt.Errorf("redis URL is required")
	}

	if c.PoolSize <= 0 {
		return fmt.Errorf("pool size must be greater than 0")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.CacheTTL <= 0 {
		return fmt.Errorf("cache TTL must be greater than 0")
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}
