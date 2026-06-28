package space

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisSpace is an implementation of Space backed by Redis.
type RedisSpace struct {
	client *redis.Client
}

// NewRedisSpace creates a new RedisSpace with the given Redis options.
func NewRedisSpace(opts *redis.Options) *RedisSpace {
	return &RedisSpace{
		client: redis.NewClient(opts),
	}
}

// Put stores a tuple into Redis, respecting TTL if set.
func (r *RedisSpace) Put(ctx context.Context, tuple Tuple) error {
	if tuple.TTL > 0 {
		return r.client.Set(ctx, tuple.Key, tuple.Value, tuple.TTL).Err()
	}
	return r.client.Set(ctx, tuple.Key, tuple.Value, 0).Err()
}

// Get retrieves a tuple by key from Redis, including remaining TTL.
func (r *RedisSpace) Get(ctx context.Context, key string) (Tuple, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return Tuple{}, err
	}
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		ttl = 0
	}
	return Tuple{Key: key, Value: data, TTL: ttl}, nil
}

// Keys returns all keys matching the given pattern.
func (r *RedisSpace) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}
