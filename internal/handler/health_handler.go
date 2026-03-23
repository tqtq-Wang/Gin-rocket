package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	serviceName string
	sqlDB       *sql.DB
	redis       *redis.Client
}

func NewHealthHandler(serviceName string, db *gorm.DB, redisClient *redis.Client) (*HealthHandler, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	return &HealthHandler{
		serviceName: serviceName,
		sqlDB:       sqlDB,
		redis:       redisClient,
	}, nil
}

func (h *HealthHandler) Live(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"service": h.serviceName,
		"status":  "ok",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	components := gin.H{}
	httpStatus := http.StatusOK

	if err := h.sqlDB.PingContext(ctx); err != nil {
		httpStatus = http.StatusServiceUnavailable
		components["mysql"] = "down"
	} else {
		components["mysql"] = "up"
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		httpStatus = http.StatusServiceUnavailable
		components["redis"] = "down"
	} else {
		components["redis"] = "up"
	}

	message := "ready"
	if httpStatus != http.StatusOK {
		message = "dependencies unavailable"
	}

	response.JSON(c, httpStatus, message, gin.H{
		"service":    h.serviceName,
		"components": components,
	})
}
