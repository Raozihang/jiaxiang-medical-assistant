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

## 登录（开发环境）
- 账号默认值：`student`、`doctor`、`admin`
- 必须修改：`AUTH_JWT_SECRET`、`AUTH_STUDENT_PASSWORD`、`AUTH_DOCTOR_PASSWORD`、`AUTH_ADMIN_PASSWORD`
- 若仍使用占位符或弱口令，服务会在启动时直接失败。
