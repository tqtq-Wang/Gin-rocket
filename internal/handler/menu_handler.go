package handler

import (
	"net/http"

	"gin-rocket/internal/middleware"
	"gin-rocket/internal/service"
	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	permissionService service.PermissionService
}

func NewMenuHandler(permissionService service.PermissionService) *MenuHandler {
	return &MenuHandler{permissionService: permissionService}
}

// CurrentUserMenus 返回当前登录用户的菜单权限树。
func (h *MenuHandler) CurrentUserMenus(c *gin.Context) {
	currentUser, ok := middleware.GetAuthUser(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	menus, err := h.permissionService.GetCurrentUserMenus(c.Request.Context(), currentUser.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "查询菜单失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"menus": menus})
}
