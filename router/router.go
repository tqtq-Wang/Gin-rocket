package router

import (
	"gin-rocket/internal/handler"
	"gin-rocket/internal/middleware"
	"gin-rocket/pkg/configx"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func New(
	cfg *configx.Config,
	log *zap.Logger,
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
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

	v1 := engine.Group("/api/v1")
	{
		users := v1.Group("/users")
		users.GET("/:id", userHandler.GetByID)
	}

	return engine
}
