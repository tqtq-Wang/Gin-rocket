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
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"Admin@123456"`
	Nickname string `json:"nickname" binding:"required" example:"admin"`
	Email    string `json:"email" example:"admin@example.com"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"Admin@123456"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"refresh-token"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register user
// @Description Create a new user. The first registered user becomes super-admin automatically.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "register request"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 409 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/v1/auth/register [post]
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

// Login godoc
// @Summary Login
// @Description Login with username and password, then return tokens, user info and permission codes.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "login request"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/v1/auth/login [post]
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

// Refresh godoc
// @Summary Refresh token
// @Description Exchange refresh token for a new access token and refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "refresh request"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/v1/auth/refresh [post]
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

// Me godoc
// @Summary Get current user
// @Description Return current user profile and permission code list for menu and button rendering.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/v1/auth/me [get]
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
