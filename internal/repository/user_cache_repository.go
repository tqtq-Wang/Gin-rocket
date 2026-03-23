package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gin-rocket/internal/model"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type UserCacheRepository interface {
	Get(ctx context.Context, id uint64) (*model.User, error)
	Set(ctx context.Context, user *model.User, ttl time.Duration) error
}

type RedisUserCacheRepository struct {
	client *redis.Client
}

func NewRedisUserCacheRepository(client *redis.Client) *RedisUserCacheRepository {
	return &RedisUserCacheRepository{client: client}
}

func (r *RedisUserCacheRepository) Get(ctx context.Context, id uint64) (*model.User, error) {
	value, err := r.client.Get(ctx, userCacheKey(id)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *RedisUserCacheRepository) Set(ctx context.Context, user *model.User, ttl time.Duration) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, userCacheKey(user.ID), payload, ttl).Err()
}

func userCacheKey(id uint64) string {
	return fmt.Sprintf("user:detail:%d", id)
}
