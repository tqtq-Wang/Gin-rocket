# Gin Rocket

基于 `Golang + Gin + GORM + MySQL + Redis + Zap + Viper` 的企业级后端脚手架，当前已经内置完整 RBAC 权限系统。

## 项目结构

```text
.
- cmd/
  - server/            程序启动入口
- config/
  - config.yaml        示例配置文件
- internal/
  - bootstrap/         应用装配、RBAC 元数据初始化
  - handler/           HTTP 处理层
  - middleware/        认证、鉴权、日志中间件
  - model/             users / roles / permissions 等模型
  - repository/        MySQL / Redis 数据访问层
  - service/           认证、用户、权限服务层
- pkg/
  - cache/             Redis 初始化
  - configx/           Viper 配置管理
  - database/          GORM / MySQL 初始化
  - jwtx/              JWT 管理
  - logger/            Zap 日志初始化
  - response/          统一响应结构
  - security/          密码加密
- router/
  - router.go          路由注册
```

## 已实现功能

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

## 权限模型说明

### 角色

- `super-admin`
  说明：系统超级管理员，默认拥有全部权限

### 权限类型

- `api`
  说明：接口权限，按 `HTTP Method + Gin 路由模板` 进行校验
- `menu`
  说明：前端菜单权限，用于返回菜单树

## 首个注册用户自动成为超管

系统启动时只会初始化：

- `super-admin` 角色
- 默认菜单权限和示例接口权限

系统不会再预置管理员账号。

第一个调用注册接口创建的用户，会在事务内自动绑定 `super-admin` 角色，避免并发下出现多个首个用户。

## 接口列表

### 公共接口

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /healthz`
- `GET /readyz`

### 登录后接口

- `GET /api/v1/auth/me`
- `GET /api/v1/menus`

### 权限校验接口

- `GET /api/v1/users/:id`

## 启动方式

1. 准备 MySQL 和 Redis。
2. 修改 [config.yaml](/C:/Code/Gin-rocket/config/config.yaml) 中的连接配置与密钥配置。
3. 启动服务：

```bash
go run ./cmd/server
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
```
