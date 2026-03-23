package handler

import (
	"errors"
	"net/http"
	"strings"

	"gin-rocket/internal/middleware"
	"gin-rocket/internal/service"
	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService service.AuthService
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	Email    string `json:"email"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 创建新用户，并在首个用户时自动授予超级管理员角色。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "注册请求参数无效")
		return
	}

	result, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Username: strings.TrimSpace(req.Username),
		Password: req.Password,
		Nickname: strings.TrimSpace(req.Nickname),
		Email:    strings.TrimSpace(req.Email),
	})
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.Fail(c, http.StatusConflict, "用户名已存在")
			return
		}

		response.Fail(c, http.StatusInternalServerError, "注册失败")
		return
	}

	response.Success(c, http.StatusCreated, result)
}

// Login 用用户名密码换取访问令牌和刷新令牌。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "登录请求参数无效")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Fail(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		response.Fail(c, http.StatusInternalServerError, "登录失败")
		return
	}

	response.Success(c, http.StatusOK, result)
}

// Refresh 使用刷新令牌换取新的 token 对。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "刷新令牌请求参数无效")
		return
	}

	tokenPair, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusUnauthorized, "刷新令牌无效")
			return
		}

		response.Fail(c, http.StatusInternalServerError, "刷新令牌失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"token_pair": tokenPair})
}

// Me 返回当前登录用户和其权限码列表。
func (h *AuthHandler) Me(c *gin.Context) {
	currentUser, ok := middleware.GetAuthUser(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	user, permissionCodes, err := h.authService.GetCurrentUser(c.Request.Context(), currentUser.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "查询当前用户失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"user":             user,
		"permission_codes": permissionCodes,
	})
}
