package middleware

import (
	"net/http"
	"strings"

	"gin-rocket/internal/service"
	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
)

const authUserContextKey = "auth_user"

type AuthUser struct {
	UserID    uint64
	Username  string
	RoleCodes []string
}

// JWTAuth 从 Bearer Token 中解析用户身份，并写入 Gin 上下文。
func JWTAuth(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, "缺少认证令牌")
			c.Abort()
			return
		}

		claims, err := authService.ParseAccessToken(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "认证令牌无效")
			c.Abort()
			return
		}

		c.Set(authUserContextKey, &AuthUser{
			UserID:    claims.UserID,
			Username:  claims.Username,
			RoleCodes: claims.RoleCodes,
		})
		c.Next()
	}
}

// GetAuthUser 获取认证中间件写入的当前用户信息。
func GetAuthUser(c *gin.Context) (*AuthUser, bool) {
	value, ok := c.Get(authUserContextKey)
	if !ok {
		return nil, false
	}

	authUser, ok := value.(*AuthUser)
	return authUser, ok
}

func extractBearerToken(authorization string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
}
