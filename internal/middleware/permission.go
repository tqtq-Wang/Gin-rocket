package middleware

import (
	"net/http"

	"gin-rocket/internal/service"
	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
)

// Permission 按照接口方法和 Gin 路由模板进行动态权限校验。
func Permission(permissionService service.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, ok := GetAuthUser(c)
		if !ok {
			response.Fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		allowed, permission, err := permissionService.CheckAccess(
			c.Request.Context(),
			currentUser.UserID,
			currentUser.RoleCodes,
			c.Request.Method,
			path,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "权限校验失败")
			c.Abort()
			return
		}

		if !allowed {
			message := "无权访问当前接口"
			if permission != nil && permission.Name != "" {
				message = "缺少权限: " + permission.Name
			}

			response.Fail(c, http.StatusForbidden, message)
			c.Abort()
			return
		}

		c.Next()
	}
}
