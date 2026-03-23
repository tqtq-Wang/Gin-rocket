# Gin Rocket

基于 `Golang + Gin + GORM + MySQL + Redis + Zap + Viper` 的企业级后端脚手架，当前已内置 RBAC、JWT、Swagger 和可选 MinIO 文件能力。

## 已实现能力

### 权限模型

- `api`
  说明：基于 `Gin 路由路径 + HTTP 方法` 控制接口权限
- `menu`
  说明：用于前端菜单渲染
- `button`
  说明：用于前端按钮级权限控制，当前以权限码列表形式返回，可继续扩展

### 认证与权限

- 用户注册
- 用户登录
- `access token + refresh token`
- 登录返回用户信息和权限列表
- RBAC 权限校验中间件
- Redis 权限缓存
- 首个注册用户自动成为 `super-admin`

### Swagger

- 认证接口已添加 Swagger 注解
- 支持 Bearer Token 认证展示
- 默认启用
- 访问地址：`/swagger/index.html`

### MinIO

- `pkg/storage` 已封装 MinIO 工具类
- 支持自动创建 bucket
- 支持通过配置定义 `endpoint`、`access_key`、`secret_key`
- 支持单文件上传
- 支持多文件上传
- 支持文件删除
- 支持预签名 URL
- 支持按业务分类存储，如 `avatar/`、`docs/`

## 文件接口

- `POST /api/v1/files/upload`
  表单字段：
  `category` 可选，如 `avatar`
  `file` 必填
- `POST /api/v1/files/upload-multiple`
  表单字段：
  `category` 可选，如 `docs`
  `files` 必填，可多文件
- `DELETE /api/v1/files`
  JSON 字段：
  `object_key`
- `GET /api/v1/files/presign?object_key=...&expires=15m`

## MinIO 配置

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
  presign_expiry: 15m
```

## 启动方式

1. 准备 MySQL 和 Redis。
2. 按需准备 MinIO。
3. 修改 `config/config.yaml`。
4. 启动服务：

```bash
go run ./cmd/server
```
