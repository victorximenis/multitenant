package multitenant

import (
	"time"
)

// ConfigBuilder provides a fluent interface for building configuration
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder with default values
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{
			DatabaseType: PostgreSQL,
			HeaderName:   "X-Tenant-Id",
			CacheTTL:     5 * time.Minute,
			PoolSize:     10,
			MaxRetries:   3,
			RetryDelay:   1 * time.Second,
			LogLevel:     "info",
		},
	}
}

// WithPostgreSQL configures PostgreSQL as the database
func (b *ConfigBuilder) WithPostgreSQL(dsn string) *ConfigBuilder {
	b.config.DatabaseType = PostgreSQL
	b.config.DatabaseDSN = dsn
	return b
}

// WithMongoDB configures MongoDB as the database
func (b *ConfigBuilder) WithMongoDB(dsn string) *ConfigBuilder {
	b.config.DatabaseType = MongoDB
	b.config.DatabaseDSN = dsn
	return b
}

// WithRedis configures Redis connection
func (b *ConfigBuilder) WithRedis(url string) *ConfigBuilder {
	b.config.RedisURL = url
	return b
}

// WithCacheTTL sets the cache TTL
func (b *ConfigBuilder) WithCacheTTL(ttl time.Duration) *ConfigBuilder {
	b.config.CacheTTL = ttl
	return b
}

// WithHeaderName sets the HTTP header name for tenant identification
func (b *ConfigBuilder) WithHeaderName(name string) *ConfigBuilder {
	b.config.HeaderName = name
	return b
}

// WithPoolSize sets the connection pool size
func (b *ConfigBuilder) WithPoolSize(size int) *ConfigBuilder {
	b.config.PoolSize = size
	return b
}

// WithRetryConfig sets retry configuration
func (b *ConfigBuilder) WithRetryConfig(maxRetries int, delay time.Duration) *ConfigBuilder {
	b.config.MaxRetries = maxRetries
	b.config.RetryDelay = delay
	return b
}

// WithLogLevel sets the log level
func (b *ConfigBuilder) WithLogLevel(level string) *ConfigBuilder {
	b.config.LogLevel = level
	return b
}

// WithDevelopmentDefaults sets development-friendly defaults
func (b *ConfigBuilder) WithDevelopmentDefaults() *ConfigBuilder {
	b.config.LogLevel = "debug"
	b.config.CacheTTL = 1 * time.Minute
	b.config.PoolSize = 5
	return b
}

// WithProductionDefaults sets production-friendly defaults
func (b *ConfigBuilder) WithProductionDefaults() *ConfigBuilder {
	b.config.LogLevel = "warn"
	b.config.CacheTTL = 15 * time.Minute
	b.config.PoolSize = 20
	b.config.MaxRetries = 5
	b.config.RetryDelay = 2 * time.Second
	return b
}

// WithTestDefaults sets test-friendly defaults
func (b *ConfigBuilder) WithTestDefaults() *ConfigBuilder {
	b.config.LogLevel = "error"
	b.config.CacheTTL = 30 * time.Second
	b.config.PoolSize = 2
	b.config.MaxRetries = 1
	b.config.RetryDelay = 100 * time.Millisecond
	return b
}

// WithIgnoredEndpoints sets the endpoints to be ignored by the middleware
func (b *ConfigBuilder) WithIgnoredEndpoints(endpoints []string) *ConfigBuilder {
	b.config.IgnoredEndpoints = endpoints
	return b
}

// WithIgnoredEndpoint adds a single endpoint to be ignored by the middleware
func (b *ConfigBuilder) WithIgnoredEndpoint(endpoint string) *ConfigBuilder {
	if b.config.IgnoredEndpoints == nil {
		b.config.IgnoredEndpoints = make([]string, 0)
	}
	b.config.IgnoredEndpoints = append(b.config.IgnoredEndpoints, endpoint)
	return b
}

// Build validates and returns the configuration
func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// MustBuild validates and returns the configuration, panicking on error
func (b *ConfigBuilder) MustBuild() *Config {
	config, err := b.Build()
	if err != nil {
		panic(err)
	}
	return config
}

// Clone creates a copy of the current configuration
func (b *ConfigBuilder) Clone() *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{
			DatabaseType:     b.config.DatabaseType,
			DatabaseDSN:      b.config.DatabaseDSN,
			RedisURL:         b.config.RedisURL,
			CacheTTL:         b.config.CacheTTL,
			HeaderName:       b.config.HeaderName,
			PoolSize:         b.config.PoolSize,
			MaxRetries:       b.config.MaxRetries,
			RetryDelay:       b.config.RetryDelay,
			LogLevel:         b.config.LogLevel,
			IgnoredEndpoints: append([]string(nil), b.config.IgnoredEndpoints...),
		},
	}
}

// FromConfig creates a builder from an existing configuration
func FromConfig(config *Config) *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{
			DatabaseType:     config.DatabaseType,
			DatabaseDSN:      config.DatabaseDSN,
			RedisURL:         config.RedisURL,
			CacheTTL:         config.CacheTTL,
			HeaderName:       config.HeaderName,
			PoolSize:         config.PoolSize,
			MaxRetries:       config.MaxRetries,
			RetryDelay:       config.RetryDelay,
			LogLevel:         config.LogLevel,
			IgnoredEndpoints: append([]string(nil), config.IgnoredEndpoints...),
		},
	}
}

// Preset configurations for common scenarios

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig(postgresURL, redisURL string) *Config {
	return NewConfigBuilder().
		WithPostgreSQL(postgresURL).
		WithRedis(redisURL).
		WithDevelopmentDefaults().
		MustBuild()
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig(postgresURL, redisURL string) *Config {
	return NewConfigBuilder().
		WithPostgreSQL(postgresURL).
		WithRedis(redisURL).
		WithProductionDefaults().
		MustBuild()
}

// TestConfig returns a configuration suitable for testing
func TestConfig(postgresURL, redisURL string) *Config {
	return NewConfigBuilder().
		WithPostgreSQL(postgresURL).
		WithRedis(redisURL).
		WithTestDefaults().
		MustBuild()
}

// MongoConfig returns a configuration using MongoDB
func MongoConfig(mongoURL, redisURL string) *Config {
	return NewConfigBuilder().
		WithMongoDB(mongoURL).
		WithRedis(redisURL).
		MustBuild()
}
