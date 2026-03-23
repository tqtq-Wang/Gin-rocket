package middleware

import (
	"net/http"
	"runtime/debug"

	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error(
			"panic recovered",
			zap.Any("panic", recovered),
			zap.ByteString("stack", debug.Stack()),
			zap.String("request_id", GetRequestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		response.Fail(c, http.StatusInternalServerError, "internal server error")
	})
}
