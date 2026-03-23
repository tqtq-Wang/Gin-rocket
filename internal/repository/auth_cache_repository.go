package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthorizationCacheRepository interface {
	GetPermissionCodes(ctx context.Context, userID uint64) ([]string, error)
	SetPermissionCodes(ctx context.Context, userID uint64, codes []string, ttl time.Duration) error
	DeletePermissionCodes(ctx context.Context, userID uint64) error
}

type RedisAuthorizationCacheRepository struct {
	client *redis.Client
}

func NewRedisAuthorizationCacheRepository(client *redis.Client) *RedisAuthorizationCacheRepository {
	return &RedisAuthorizationCacheRepository{client: client}
}

func (r *RedisAuthorizationCacheRepository) GetPermissionCodes(ctx context.Context, userID uint64) ([]string, error) {
	payload, err := r.client.Get(ctx, permissionCodesCacheKey(userID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var codes []string
	if err := json.Unmarshal([]byte(payload), &codes); err != nil {
		return nil, err
	}

	return codes, nil
}

func (r *RedisAuthorizationCacheRepository) SetPermissionCodes(ctx context.Context, userID uint64, codes []string, ttl time.Duration) error {
	payload, err := json.Marshal(codes)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, permissionCodesCacheKey(userID), payload, ttl).Err()
}

func (r *RedisAuthorizationCacheRepository) DeletePermissionCodes(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, permissionCodesCacheKey(userID)).Err()
}

func permissionCodesCacheKey(userID uint64) string {
	return fmt.Sprintf("rbac:user:permissions:%d", userID)
}
