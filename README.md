# Gin Rocket

基于 `Golang + Gin + GORM + MySQL + Redis + Zap + Viper` 的企业级后端脚手架，当前已内置完整 RBAC、JWT 认证、Swagger 文档和可选 MinIO 集成。

## 当前权限模型

已支持的权限类型：

- `api`
  说明：基于 `Gin 路由路径 + HTTP 方法` 进行权限控制
- `menu`
  说明：用于前端动态菜单渲染
- `button`
  说明：用于前端按钮级权限控制，当前通过权限码列表下发，后续可继续扩展更细粒度的按钮模型

## 已实现功能

### 认证与权限

- 用户注册
- 用户登录
- JWT 认证，包含 `access token + refresh token`
- 登录返回用户信息和权限列表
- RBAC 权限校验中间件，自动拦截无权限请求
- Redis 权限缓存
- 首个注册用户自动成为 `super-admin`

### Swagger

- 认证接口已补充 Swagger 注解
- 支持 Bearer Token 认证展示
- 默认启用
- 访问地址：`/swagger/index.html`

### MinIO

- 面向开发者按需启用
- 默认关闭
- 启用后自动初始化客户端
- 可按配置自动创建桶

## 关键接口

### 认证接口

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`

### 权限相关接口

- `GET /api/v1/menus`
- `GET /api/v1/users/:id`

## 配置说明

### Swagger

```yaml
swagger:
  enabled: true
  route_prefix: /swagger
  title: Gin Rocket API
  description: Gin Rocket backend APIs
  version: 1.0.0
  server_url: ""
```

### MinIO

```yaml
minio:
  enabled: false
  endpoint: 127.0.0.1:9000
  access_key: minioadmin
  secret_key: minioadmin
  use_ssl: false
  bucket: gin-rocket
  location: ap-east-1
  auto_create_bucket: true
  base_url: ""
```

## 启动方式

1. 准备 MySQL 和 Redis。
2. 按需准备 MinIO。
3. 修改 `config/config.yaml`。
4. 启动服务：

```bash
go run ./cmd/server
```
