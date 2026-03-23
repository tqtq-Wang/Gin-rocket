package jwtx

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gin-rocket/internal/model"
	"gin-rocket/pkg/configx"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var ErrInvalidTokenType = errors.New("invalid token type")

type Claims struct {
	UserID    uint64   `json:"user_id"`
	Username  string   `json:"username"`
	RoleCodes []string `json:"role_codes"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	ExpiresIn        int64     `json:"expires_in"`
}

type Manager struct {
	issuer        string
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(cfg configx.AuthConfig) *Manager {
	return &Manager{
		issuer:        cfg.Issuer,
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

// GenerateTokenPair 同时生成访问令牌和刷新令牌，刷新令牌包含唯一 jti 用于轮换控制。
func (m *Manager) GenerateTokenPair(user *model.User) (*TokenPair, *Claims, error) {
	now := time.Now()
	roleCodes := extractRoleCodes(user.Roles)
	accessExpiresAt := now.Add(m.accessTTL)
	refreshExpiresAt := now.Add(m.refreshTTL)

	accessClaims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		RoleCodes: roleCodes,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatUint(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshClaims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		RoleCodes: roleCodes,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    m.issuer,
			Subject:   strconv.FormatUint(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken, err := m.signToken(accessClaims, m.accessSecret)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := m.signToken(refreshClaims, m.refreshSecret)
	if err != nil {
		return nil, nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		ExpiresIn:        int64(m.accessTTL.Seconds()),
	}, refreshClaims, nil
}

// ParseAccessToken 只接受 access token。
func (m *Manager) ParseAccessToken(token string) (*Claims, error) {
	claims, err := m.parseToken(token, m.accessSecret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ParseRefreshToken 只接受 refresh token。
func (m *Manager) ParseRefreshToken(token string) (*Claims, error) {
	claims, err := m.parseToken(token, m.refreshSecret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func (m *Manager) signToken(claims *Claims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func (m *Manager) parseToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func extractRoleCodes(roles []model.Role) []string {
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleCodes = append(roleCodes, role.Code)
	}

	return roleCodes
}
