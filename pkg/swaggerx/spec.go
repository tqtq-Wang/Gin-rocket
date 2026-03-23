package swaggerx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	texttemplate "text/template"

	"gin-rocket/pkg/configx"
)

var openAPISpecTemplate = texttemplate.Must(texttemplate.New("openapi").Parse(openapiYAMLTemplate))
var swaggerUITemplate = template.Must(template.New("swagger-ui").Parse(swaggerUIHTMLTemplate))

type specData struct {
	Title       string
	Description string
	Version     string
	ServerURL   string
}

// ResolveServerURL uses configured server URL first and falls back to the current request host.
func ResolveServerURL(cfg configx.SwaggerConfig, r *http.Request) string {
	if cfg.ServerURL != "" {
		return strings.TrimRight(cfg.ServerURL, "/")
	}

	scheme := "http"
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	} else if r.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// RenderSpec generates the OpenAPI YAML payload from runtime configuration.
func RenderSpec(cfg configx.SwaggerConfig, serverURL string) ([]byte, error) {
	var buf bytes.Buffer
	if err := openAPISpecTemplate.Execute(&buf, specData{
		Title:       cfg.Title,
		Description: cfg.Description,
		Version:     cfg.Version,
		ServerURL:   serverURL,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RenderUI renders a simple Swagger UI page for the generated OpenAPI spec.
func RenderUI(title, specURL string) ([]byte, error) {
	var buf bytes.Buffer
	if err := swaggerUITemplate.Execute(&buf, map[string]string{
		"title":   title,
		"specURL": specURL,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

const openapiYAMLTemplate = `openapi: 3.0.3
info:
  title: {{ .Title }}
  description: {{ .Description }}
  version: {{ .Version }}
servers:
  - url: {{ .ServerURL }}
paths:
  /healthz:
    get:
      summary: 存活检查
      responses:
        '200':
          description: OK
  /readyz:
    get:
      summary: 就绪检查
      responses:
        '200':
          description: Ready
  /api/v1/auth/register:
    post:
      summary: 用户注册
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegisterRequest'
      responses:
        '201':
          description: Created
  /api/v1/auth/login:
    post:
      summary: 用户登录
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: OK
  /api/v1/auth/refresh:
    post:
      summary: 刷新令牌
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RefreshRequest'
      responses:
        '200':
          description: OK
  /api/v1/auth/me:
    get:
      summary: 获取当前用户
      security:
        - bearerAuth: []
      responses:
        '200':
          description: OK
  /api/v1/menus:
    get:
      summary: 获取当前用户菜单
      security:
        - bearerAuth: []
      responses:
        '200':
          description: OK
  /api/v1/users/{id}:
    get:
      summary: 获取用户详情
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  schemas:
    RegisterRequest:
      type: object
      required: [username, password, nickname]
      properties:
        username:
          type: string
        password:
          type: string
        nickname:
          type: string
        email:
          type: string
    LoginRequest:
      type: object
      required: [username, password]
      properties:
        username:
          type: string
        password:
          type: string
    RefreshRequest:
      type: object
      required: [refresh_token]
      properties:
        refresh_token:
          type: string
`

const swaggerUIHTMLTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{{ .title }}</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "{{ .specURL }}",
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>
`
