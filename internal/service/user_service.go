package service

import (
	"context"

	"gin-rocket/internal/model"
	"gin-rocket/internal/repository"
)

type UserService interface {
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetProfile(ctx context.Context, id uint64) (*model.User, error)
}

type DefaultUserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *DefaultUserService {
	return &DefaultUserService{userRepo: userRepo}
}

func (s *DefaultUserService) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *DefaultUserService) GetProfile(ctx context.Context, id uint64) (*model.User, error) {
	return s.userRepo.GetProfileByID(ctx, id)
}
