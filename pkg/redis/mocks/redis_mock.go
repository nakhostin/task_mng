package mocks

import (
	"context"
	"time"
)

// MockRedisClient is a mock implementation of redis.RedisClient
type MockRedisClient struct {
	SetFunc         func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetFunc         func(ctx context.Context, key string) (string, error)
	DelFunc         func(ctx context.Context, keys ...string) error
	ExistsFunc      func(ctx context.Context, keys ...string) (int64, error)
	ExpireFunc      func(ctx context.Context, key string, expiration time.Duration) error
	TTLFunc         func(ctx context.Context, key string) (time.Duration, error)
	IncrFunc        func(ctx context.Context, key string) error
	IncrByFunc      func(ctx context.Context, key string, value int64) error
	HealthCheckFunc func() error
	CloseFunc       func() error
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, expiration)
	}
	return nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, keys...)
	}
	return nil
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, keys...)
	}
	return 0, nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if m.ExpireFunc != nil {
		return m.ExpireFunc(ctx, key, expiration)
	}
	return nil
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if m.TTLFunc != nil {
		return m.TTLFunc(ctx, key)
	}
	return 0, nil
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) error {
	if m.IncrFunc != nil {
		return m.IncrFunc(ctx, key)
	}
	return nil
}

func (m *MockRedisClient) IncrBy(ctx context.Context, key string, value int64) error {
	if m.IncrByFunc != nil {
		return m.IncrByFunc(ctx, key, value)
	}
	return nil
}

func (m *MockRedisClient) HealthCheck() error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc()
	}
	return nil
}

func (m *MockRedisClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
