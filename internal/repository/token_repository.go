package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenRepository interface {
	Store(ctx context.Context, userID uint64, tokenID string, ttl time.Duration) error
	Exists(ctx context.Context, userID uint64, tokenID string) (bool, error)
	Delete(ctx context.Context, userID uint64, tokenID string) error
}

type RedisRefreshTokenRepository struct {
	client *redis.Client
	prefix string
}

func NewRedisRefreshTokenRepository(client *redis.Client, prefix string) *RedisRefreshTokenRepository {
	return &RedisRefreshTokenRepository{
		client: client,
		prefix: prefix,
	}
}

func (r *RedisRefreshTokenRepository) Store(ctx context.Context, userID uint64, tokenID string, ttl time.Duration) error {
	return r.client.Set(ctx, refreshTokenKey(r.prefix, userID, tokenID), 1, ttl).Err()
}

func (r *RedisRefreshTokenRepository) Exists(ctx context.Context, userID uint64, tokenID string) (bool, error) {
	count, err := r.client.Exists(ctx, refreshTokenKey(r.prefix, userID, tokenID)).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *RedisRefreshTokenRepository) Delete(ctx context.Context, userID uint64, tokenID string) error {
	return r.client.Del(ctx, refreshTokenKey(r.prefix, userID, tokenID)).Err()
}

func refreshTokenKey(prefix string, userID uint64, tokenID string) string {
	return fmt.Sprintf("%s:%d:%s", prefix, userID, tokenID)
}
