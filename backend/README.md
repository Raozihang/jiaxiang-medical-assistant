# Backend（Go + Gin）

## 启动
```bash
go mod tidy
go run ./cmd/server
```

## 目录
- `cmd/server`：服务启动入口
- `internal/bootstrap`：配置、数据库与路由装配
- `internal/handler`：HTTP 接口处理
- `internal/model`：数据库模型定义
- `internal/middleware`：请求中间件
- `internal/response`：统一响应结构

## 环境变量
复制 `.env.example` 到 `.env` 后按需修改数据库连接信息。
