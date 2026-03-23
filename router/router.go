package router

import (
	"gin-rocket/internal/handler"
	"gin-rocket/internal/middleware"
	"gin-rocket/internal/service"
	"gin-rocket/pkg/configx"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func New(
	cfg *configx.Config,
	log *zap.Logger,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	menuHandler *handler.MenuHandler,
	userHandler *handler.UserHandler,
	authService service.AuthService,
	permissionService service.PermissionService,
) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.RequestLogger(log),
		middleware.Recovery(log),
	)

	engine.GET("/healthz", healthHandler.Live)
	engine.GET("/readyz", healthHandler.Ready)

	if cfg.Swagger.Enabled {
		swaggerPath := strings.TrimRight(cfg.Swagger.RoutePrefix, "/")
		engine.GET(swaggerPath, func(c *gin.Context) {
			c.Redirect(302, swaggerPath+"/index.html")
		})
		engine.GET(swaggerPath+"/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	v1 := engine.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)

		authorized := v1.Group("")
		authorized.Use(middleware.JWTAuth(authService))
		authorized.GET("/auth/me", authHandler.Me)
		authorized.GET("/menus", menuHandler.CurrentUserMenus)

		permissionProtected := authorized.Group("")
		permissionProtected.Use(middleware.Permission(permissionService))
		permissionProtected.GET("/users/:id", userHandler.GetByID)
	}

	return engine
}
