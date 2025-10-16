package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Incr(ctx context.Context, key string) error
	IncrBy(ctx context.Context, key string, value int64) error
	HealthCheck() error
	Close() error
}

// Redis represents a Redis client
type Redis struct {
	client *redis.Client
}

// New creates a new Redis client
func New(cfg Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	slog.Info("Redis connection established",
		"host", cfg.Host,
		"port", cfg.Port,
		"db", cfg.DB)

	return &Redis{client: client}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	slog.Info("Closing Redis connection")
	return r.client.Close()
}

// Set sets a key-value pair with expiration
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del deletes a key
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL gets the time to live for a key
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Incr increments a key
func (r *Redis) Incr(ctx context.Context, key string) error {
	return r.client.Incr(ctx, key).Err()
}

// IncrBy increments a key by a specific amount
func (r *Redis) IncrBy(ctx context.Context, key string, value int64) error {
	return r.client.IncrBy(ctx, key, value).Err()
}

// HealthCheck checks if Redis is healthy
func (r *Redis) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Ping(ctx).Err()
}
