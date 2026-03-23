package service

import (
	"context"
	"errors"
	"time"

	"gin-rocket/internal/model"
	"gin-rocket/internal/repository"
	"gin-rocket/pkg/jwtx"
	"gin-rocket/pkg/security"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type LoginResult struct {
	TokenPair       *jwtx.TokenPair `json:"token_pair"`
	User            *model.User     `json:"user"`
	PermissionCodes []string        `json:"permission_codes"`
	IsSuperAdmin    bool            `json:"is_super_admin"`
}

type RegisterInput struct {
	Username string
	Password string
	Nickname string
	Email    string
}

// AuthService 负责注册、登录、刷新令牌和当前用户查询。
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*LoginResult, error)
	Login(ctx context.Context, username, password string) (*LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (*jwtx.TokenPair, error)
	ParseAccessToken(token string) (*jwtx.Claims, error)
	GetCurrentUser(ctx context.Context, userID uint64) (*model.User, []string, error)
}

type DefaultAuthService struct {
	userRepo               repository.UserRepository
	refreshTokenRepo       repository.RefreshTokenRepository
	authorizationCacheRepo repository.AuthorizationCacheRepository
	jwtManager             *jwtx.Manager
	permissionCacheTTL     time.Duration
	logger                 *zap.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	authorizationCacheRepo repository.AuthorizationCacheRepository,
	jwtManager *jwtx.Manager,
	permissionCacheTTL time.Duration,
	logger *zap.Logger,
) *DefaultAuthService {
	return &DefaultAuthService{
		userRepo:               userRepo,
		refreshTokenRepo:       refreshTokenRepo,
		authorizationCacheRepo: authorizationCacheRepo,
		jwtManager:             jwtManager,
		permissionCacheTTL:     permissionCacheTTL,
		logger:                 logger,
	}
}

// Register 创建新用户，并在首用户场景下自动授予 super-admin 角色。
func (s *DefaultAuthService) Register(ctx context.Context, input RegisterInput) (*LoginResult, error) {
	hashedPassword, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: input.Username,
		Password: hashedPassword,
		Nickname: input.Nickname,
		Email:    input.Email,
		Status:   model.StatusEnabled,
	}

	isSuperAdmin, err := s.userRepo.Register(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	tokenPair, refreshClaims, err := s.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokenRepo.Store(ctx, user.ID, refreshClaims.ID, time.Until(tokenPair.RefreshExpiresAt)); err != nil {
		return nil, err
	}

	permissionCodes, err := s.loadPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		TokenPair:       tokenPair,
		User:            user,
		PermissionCodes: permissionCodes,
		IsSuperAdmin:    isSuperAdmin,
	}, nil
}

// Login 校验账号密码，并签发新的访问令牌和刷新令牌。
func (s *DefaultAuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := security.ComparePassword(user.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenPair, refreshClaims, err := s.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokenRepo.Store(ctx, user.ID, refreshClaims.ID, time.Until(tokenPair.RefreshExpiresAt)); err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateLastLoginAt(ctx, user.ID, time.Now()); err != nil {
		s.logger.Warn("update last login time failed", zap.Uint64("user_id", user.ID), zap.Error(err))
	}

	permissionCodes, err := s.loadPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		TokenPair:       tokenPair,
		User:            user,
		PermissionCodes: permissionCodes,
		IsSuperAdmin:    hasRoleCodeFromUser(user),
	}, nil
}

// Refresh 校验刷新令牌是否仍在 Redis 白名单中，并轮换生成新的 token 对。
func (s *DefaultAuthService) Refresh(ctx context.Context, refreshToken string) (*jwtx.TokenPair, error) {
	claims, err := s.jwtManager.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	exists, err := s.refreshTokenRepo.Exists(ctx, claims.UserID, claims.ID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	tokenPair, newRefreshClaims, err := s.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokenRepo.Delete(ctx, claims.UserID, claims.ID); err != nil {
		return nil, err
	}

	if err := s.refreshTokenRepo.Store(ctx, claims.UserID, newRefreshClaims.ID, time.Until(tokenPair.RefreshExpiresAt)); err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *DefaultAuthService) ParseAccessToken(token string) (*jwtx.Claims, error) {
	return s.jwtManager.ParseAccessToken(token)
}

func (s *DefaultAuthService) GetCurrentUser(ctx context.Context, userID uint64) (*model.User, []string, error) {
	user, err := s.userRepo.GetProfileByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	permissionCodes, err := s.loadPermissionCodes(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, permissionCodes, nil
}

// loadPermissionCodes 优先读 Redis，未命中时回源数据库并回填缓存。
func (s *DefaultAuthService) loadPermissionCodes(ctx context.Context, userID uint64) ([]string, error) {
	codes, err := s.authorizationCacheRepo.GetPermissionCodes(ctx, userID)
	if err == nil {
		return codes, nil
	}
	if err != nil && !errors.Is(err, repository.ErrCacheMiss) {
		s.logger.Warn("load permission codes from cache failed", zap.Uint64("user_id", userID), zap.Error(err))
	}

	codes, err = s.userRepo.ListPermissionCodesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizationCacheRepo.SetPermissionCodes(ctx, userID, codes, s.permissionCacheTTL); err != nil {
		s.logger.Warn("write permission codes to cache failed", zap.Uint64("user_id", userID), zap.Error(err))
	}

	return codes, nil
}

func hasRoleCodeFromUser(user *model.User) bool {
	for _, role := range user.Roles {
		if role.Code == model.RoleCodeSuperAdmin {
			return true
		}
	}

	return false
}
