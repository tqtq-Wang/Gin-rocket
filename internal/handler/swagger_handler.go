package handler

import (
	"net/http"
	"strings"

	"gin-rocket/pkg/configx"
	"gin-rocket/pkg/swaggerx"

	"github.com/gin-gonic/gin"
)

type SwaggerHandler struct {
	cfg configx.SwaggerConfig
}

func NewSwaggerHandler(cfg configx.SwaggerConfig) *SwaggerHandler {
	return &SwaggerHandler{cfg: cfg}
}

// UI renders the Swagger UI page.
func (h *SwaggerHandler) UI(c *gin.Context) {
	specURL := strings.TrimRight(h.cfg.RoutePrefix, "/") + "/openapi.yaml"
	payload, err := swaggerx.RenderUI(h.cfg.Title, specURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "render swagger ui failed")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", payload)
}

// Spec renders the OpenAPI specification using the current runtime host.
func (h *SwaggerHandler) Spec(c *gin.Context) {
	serverURL := swaggerx.ResolveServerURL(h.cfg, c.Request)
	payload, err := swaggerx.RenderSpec(h.cfg, serverURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "render openapi spec failed")
		return
	}

	c.Data(http.StatusOK, "application/yaml; charset=utf-8", payload)
}
