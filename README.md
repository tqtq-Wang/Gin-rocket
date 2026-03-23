# Gin Rocket

基于 `Golang + Gin + GORM + MySQL + Redis + Zap + Viper` 的企业级后端脚手架，目标是结构清晰、初始化链路明确、便于长期扩展和维护。

## 目录结构

```text
.
├── cmd/
│   └── server/           # 程序启动入口
├── config/
│   └── config.yaml       # 示例配置
├── internal/
│   ├── bootstrap/        # 应用装配与资源初始化
│   ├── handler/          # HTTP Handler
│   ├── middleware/       # Gin 中间件
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层（MySQL / Redis）
│   └── service/          # 业务逻辑层
├── pkg/
│   ├── cache/            # Redis 初始化
│   ├── configx/          # Viper 配置加载
│   ├── database/         # MySQL / GORM 初始化
│   ├── logger/           # Zap 日志初始化
│   └── response/         # 统一响应结构
└── router/
    └── router.go         # 路由注册
```

## 已实现能力

- 优雅启动与关闭
- Viper 配置加载与环境变量覆盖
- Zap 日志初始化
- GORM MySQL 连接初始化
- Redis 初始化与连通性检测
- 请求 ID、访问日志、Panic Recovery 中间件
- `handler -> service -> repository -> model` 示例链路
- 健康检查与就绪检查接口

## 启动方式

1. 准备 MySQL 和 Redis。
2. 修改 `config/config.yaml` 中的连接信息。
3. 启动服务：

```bash
go run ./cmd/server
```

## 环境变量覆盖

Viper 已启用环境变量覆盖，规则为将配置键中的 `.` 替换成 `_`。

例如：

```bash
APP_NAME=gin-rocket
SERVER_PORT=8080
MYSQL_HOST=127.0.0.1
REDIS_ADDR=127.0.0.1:6379
```

也可以通过 `CONFIG_FILE` 指定外部配置文件路径：

```bash
CONFIG_FILE=./config/config.yaml
```

## 默认路由

- `GET /healthz`：存活检查
- `GET /readyz`：依赖就绪检查
- `GET /api/v1/users/:id`：示例用户查询接口
