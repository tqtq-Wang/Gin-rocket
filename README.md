# Gin Rocket

基于 `Golang + Gin + GORM + MySQL + Redis + Zap + Viper` 的企业级后端脚手架，当前已内置 RBAC、Swagger 和可选 MinIO 集成。

## 项目结构

```text
.
├── cmd/
│   └── server/            # 程序启动入口
├── config/
│   └── config.yaml        # 示例配置文件
├── internal/
│   ├── bootstrap/         # 应用装配、RBAC 元数据初始化
│   ├── handler/           # HTTP 处理层
│   ├── middleware/        # 认证、鉴权、日志中间件
│   ├── model/             # users / roles / permissions 等模型
│   ├── repository/        # MySQL / Redis 数据访问层
│   └── service/           # 认证、用户、权限服务层
├── pkg/
│   ├── cache/             # Redis 初始化
│   ├── configx/           # Viper 配置管理
│   ├── database/          # GORM / MySQL 初始化
│   ├── jwtx/              # JWT 管理
│   ├── logger/            # Zap 日志初始化
│   ├── response/          # 统一响应结构
│   ├── security/          # 密码加密
│   ├── storage/           # MinIO 客户端初始化
│   └── swaggerx/          # OpenAPI 与 Swagger UI 渲染
└── router/
    └── router.go          # 路由注册
```

## 已实现能力

### 基础能力

- 优雅启动和关闭
- 配置文件加载
- Zap 日志初始化
- MySQL 连接初始化
- Redis 初始化

### RBAC 能力

- 用户表 `users`
- 角色表 `roles`
- 权限表 `permissions`
- 用户角色关联表 `user_roles`
- 角色权限关联表 `role_permissions`
- JWT 登录
- Refresh Token 轮换
- 基于接口路径的权限控制
- 动态权限加载
- 菜单权限支持
- 首个注册用户自动成为超管

### Swagger

- 默认启用
- 访问地址：`/swagger`
- OpenAPI 文档地址：`/swagger/openapi.yaml`
- 可通过配置修改标题、描述、版本和路由前缀

### MinIO

- 默认关闭
- 仅在开发者需要对象存储能力时启用
- 启用后会初始化客户端，并按配置自动创建桶

## 首个注册用户自动成为超管

系统启动时只初始化：

- `super-admin` 角色
- 默认菜单权限
- 示例接口权限

系统不再预置管理员账号。

第一个调用注册接口创建的用户，会在事务内自动绑定 `super-admin` 角色，避免并发下出现多个首个用户。

## 接口列表

### 公共接口

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /healthz`
- `GET /readyz`
- `GET /swagger`

### 登录后接口

- `GET /api/v1/auth/me`
- `GET /api/v1/menus`

### 权限校验接口

- `GET /api/v1/users/:id`

## 启动方式

1. 准备 MySQL 和 Redis。
2. 按需准备 MinIO。
3. 修改 [config/config.yaml](/C:/Code/Gin-rocket/config/config.yaml)。
4. 启动服务：

```bash
go run ./cmd/server
```

## 关键配置

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

## 配置覆盖

Viper 已启用环境变量覆盖，规则是将配置键中的 `.` 替换为 `_`。

例如：

```bash
SERVER_PORT=8080
MYSQL_HOST=127.0.0.1
REDIS_ADDR=127.0.0.1:6379
AUTH_ACCESS_SECRET=change-me
AUTH_REFRESH_SECRET=change-me-too
MINIO_ENABLED=true
MINIO_ENDPOINT=127.0.0.1:9000
```
