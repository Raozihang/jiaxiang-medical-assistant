# 部署指南

## 1. 部署方式

### 1.1 本地部署（开发环境）

```bash
# 前端
cd frontend
npm run dev  # http://localhost:5173

# 后端
cd backend
go run cmd/server/main.go  # http://localhost:8080

# 数据库
# PostgreSQL: localhost:5432
# Redis: localhost:6379
```

### 1.2 服务器部署（生产环境）

```bash
# 1. 构建前端
cd frontend
npm run build
# 输出到 dist/ 目录

# 2. 构建后端
cd backend
go build -o server cmd/server/main.go

# 3. 配置 Nginx
sudo cp nginx.conf /etc/nginx/sites-available/medical-assistant
sudo ln -s /etc/nginx/sites-available/medical-assistant /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 4. 启动后端服务
nohup ./server > server.log 2>&1 &

# 5. 验证部署
curl https://your-domain.com/api/health
```

## 2. Docker 部署

### 2.1 Dockerfile

**前端 Dockerfile:**
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**后端 Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./server"]
```

### 2.2 Docker Compose

```yaml
version: '3.8'

services:
  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/medical_assistant
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
      - POSTGRES_DB=medical_assistant
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 2.3 启动服务

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

## 3. 环境变量

### 3.1 前端环境变量

```bash
# .env.production
VITE_API_BASE_URL=https://api.your-domain.com
VITE_APP_TITLE=嘉祥智能医务室助手
```

### 3.2 后端环境变量

```bash
# config.yaml
server:
  port: 8080
  mode: release

database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: medical_assistant

redis:
  host: localhost
  port: 6379
  password: ""

jwt:
  secret: your_secret_key
  expire: 7200  # 2 hours
```

## 4. 监控与日志

### 4.1 应用日志

```go
// 后端日志配置
log.SetFormatter(&log.JSONFormatter{})
log.SetOutput(io.MultiWriter(os.Stdout, logFile))
log.SetLevel(log.InfoLevel)
```

### 4.2 日志轮转

```bash
# 使用 logrotate
/var/log/medical-assistant/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

### 4.3 健康检查

```bash
# 健康检查接口
curl http://localhost:8080/api/health

# 响应示例
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "timestamp": 1708646400
}
```

## 5. 备份策略

### 5.1 数据库备份

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/medical_assistant"

# 备份数据库
pg_dump -U postgres medical_assistant > $BACKUP_DIR/db_$DATE.sql

# 压缩备份
gzip $BACKUP_DIR/db_$DATE.sql

# 删除 7 天前的备份
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +7 -delete
```

### 5.2 定时备份

```bash
# crontab -e
0 2 * * * /path/to/backup.sh
```

## 6. 故障排查

### Q1: 服务无法访问？
```bash
# 检查服务状态
systemctl status nginx
ps aux | grep server

# 检查端口
netstat -tlnp | grep :80
netstat -tlnp | grep :8080

# 检查防火墙
sudo ufw status
```

### Q2: 数据库连接失败？
```bash
# 检查 PostgreSQL 状态
systemctl status postgresql

# 检查连接
psql -U postgres -d medical_assistant -c "SELECT 1"
```

### Q3: 前端页面空白？
```bash
# 检查 Nginx 配置
nginx -t

# 检查前端文件
ls -la /usr/share/nginx/html/

# 查看浏览器控制台错误
```

---

*最后更新：2026-02-23*
*维护者：小饶 (Qwen)*
