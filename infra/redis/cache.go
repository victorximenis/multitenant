package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/victorximenis/multitenant/core"
)

const (
	DEFAULT_TTL = 5 * time.Minute
	KEY_PREFIX  = "multitenant:tenants:"
)

type TenantCache struct {
	client *redis.Client
	ttl    time.Duration
}

type Config struct {
	RedisURL string
	TTL      time.Duration
}

func NewTenantCache(ctx context.Context, config Config) (*TenantCache, error) {
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	ttl := config.TTL
	if ttl <= 0 {
		ttl = DEFAULT_TTL
	}

	return &TenantCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *TenantCache) tenantKey(name string) string {
	return fmt.Sprintf("%s%s", KEY_PREFIX, name)
}

func (c *TenantCache) Get(ctx context.Context, name string) (*core.Tenant, error) {
	key := c.tenantKey(name)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, core.TenantNotFoundError{Name: name}
		}
		return nil, err
	}

	var tenant core.Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (c *TenantCache) Set(ctx context.Context, tenant *core.Tenant, ttl time.Duration) error {
	if tenant == nil {
		return fmt.Errorf("tenant cannot be nil")
	}

	key := c.tenantKey(tenant.Name)

	data, err := json.Marshal(tenant)
	if err != nil {
		return err
	}

	if ttl <= 0 {
		ttl = c.ttl
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *TenantCache) Delete(ctx context.Context, name string) error {
	key := c.tenantKey(name)
	return c.client.Del(ctx, key).Err()
}

// DeleteAll removes all tenant keys from cache
func (c *TenantCache) DeleteAll(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", KEY_PREFIX)

	var cursor uint64
	var keys []string
	var err error

	for {
		keys, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
