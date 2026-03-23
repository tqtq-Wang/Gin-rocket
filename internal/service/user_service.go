package service

import (
	"context"
	"errors"
	"time"

	"gin-rocket/internal/model"
	"gin-rocket/internal/repository"

	"go.uber.org/zap"
)

type UserService interface {
	GetByID(ctx context.Context, id uint64) (*model.User, error)
}

type DefaultUserService struct {
	userRepo      repository.UserRepository
	userCacheRepo repository.UserCacheRepository
	logger        *zap.Logger
	cacheTTL      time.Duration
}

func NewUserService(
	userRepo repository.UserRepository,
	userCacheRepo repository.UserCacheRepository,
	logger *zap.Logger,
	cacheTTL time.Duration,
) *DefaultUserService {
	return &DefaultUserService{
		userRepo:      userRepo,
		userCacheRepo: userCacheRepo,
		logger:        logger,
		cacheTTL:      cacheTTL,
	}
}

func (s *DefaultUserService) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	user, err := s.userCacheRepo.Get(ctx, id)
	if err == nil {
		return user, nil
	}

	if err != nil && !errors.Is(err, repository.ErrCacheMiss) {
		s.logger.Warn("load user from cache failed", zap.Uint64("user_id", id), zap.Error(err))
	}

	user, err = s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.userCacheRepo.Set(ctx, user, s.cacheTTL); err != nil {
		s.logger.Warn("write user cache failed", zap.Uint64("user_id", id), zap.Error(err))
	}

	return user, nil
}
