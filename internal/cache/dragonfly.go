package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/northstack/platform/internal/config"
	"github.com/redis/go-redis/v9"
)

type DragonflyDB struct {
	client redis.UniversalClient
	config config.DragonflyDBConfig
}

func NewDragonflyDB(cfg config.DragonflyDBConfig) (*DragonflyDB, error) {
	var client redis.UniversalClient

	if cfg.ClusterMode {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.Addresses,
			Password: cfg.Password,
			PoolSize: cfg.PoolSize,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr(),
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
		})
	}

	// Warm up connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to dragonflydb: %w", err)
	}

	return &DragonflyDB{
		client: client,
		config: cfg,
	}, nil
}

func (d *DragonflyDB) Health(ctx context.Context) error {
	return d.client.Ping(ctx).Err()
}

func (d *DragonflyDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return d.client.Set(ctx, d.config.KeyPrefix+":"+key, data, expiration).Err()
}

func (d *DragonflyDB) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := d.client.Get(ctx, d.config.KeyPrefix+":"+key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (d *DragonflyDB) Delete(ctx context.Context, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = d.config.KeyPrefix + ":" + k
	}
	return d.client.Del(ctx, prefixedKeys...).Err()
}

// GetOrSet implements cache-aside pattern
func (d *DragonflyDB) GetOrSet(ctx context.Context, key string, dest interface{}, expiration time.Duration, fn func() (interface{}, error)) error {
	err := d.Get(ctx, key, dest)
	if err == nil {
		return nil // Cache hit
	}

	if err != redis.Nil {
		return err // Real error
	}

	// Cache miss
	val, err := fn()
	if err != nil {
		return err
	}

	if err := d.Set(ctx, key, val, expiration); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	// Copy value back to dest
	// In a real generic implementation we'd reflect, but for now we rely on the caller or this simple unmarshal
	// Optimization: Since we already marshal in Set, we can unmarshal that same byte slice
	data, _ := json.Marshal(val) // Re-marshal to be safe or optimize
	return json.Unmarshal(data, dest)
}

func (d *DragonflyDB) Close() error {
	return d.client.Close()
}
