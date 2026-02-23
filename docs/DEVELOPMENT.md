# 开发指南

## 1. 环境搭建

### 1.1 前置要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Node.js | 20.x | 前端开发 |
| Go | 1.21+ | 后端开发 |
| PostgreSQL | 15+ | 数据库 |
| Redis | 7+ | 缓存 |
| Git | 2.x | 版本控制 |

### 1.2 前端环境

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 配置环境变量
cp .env.example .env

# 启动开发服务器
npm run dev

# 访问 http://localhost:5173
```

### 1.3 后端环境

```bash
# 进入后端目录
cd backend

# 初始化 Go module
go mod init medical-assistant

# 安装依赖
go mod tidy

# 配置数据库
# 编辑 configs/config.yaml

# 启动开发服务器
go run cmd/server/main.go

# 访问 http://localhost:8080
```

### 1.4 数据库初始化

```bash
# 创建数据库
psql -U postgres -c "CREATE DATABASE medical_assistant;"

# 运行迁移
cd backend
go run cmd/migrate/main.go

# 验证连接
psql -U postgres -d medical_assistant -c "\dt"
```

## 2. 项目结构

### 2.1 前端目录

```
frontend/
├── src/
│   ├── app/              # 应用入口
│   │   ├── App.tsx
│   │   ├── routes.tsx    # 路由配置
│   │   └── store.ts      # 全局状态
│   ├── modules/          # 业务模块
│   │   ├── patient/      # 患者端
│   │   ├── doctor/       # 医生端
│   │   └── admin/        # 管理后台
│   ├── shared/           # 共享模块
│   │   ├── components/   # 通用组件
│   │   ├── hooks/        # 自定义 Hooks
│   │   ├── utils/        # 工具函数
│   │   └── api/          # API 请求
│   └── types/            # TypeScript 类型
├── public/
├── package.json
├── tsconfig.json
├── vite.config.ts
└── .env.example
```

### 2.2 后端目录

```
backend/
├── cmd/
│   ├── server/           # 应用入口
│   │   └── main.go
│   └── migrate/          # 数据库迁移
│       └── main.go
├── internal/
│   ├── handler/          # HTTP 处理器
│   ├── service/          # 业务逻辑
│   ├── repository/       # 数据访问
│   └── middleware/       # 中间件
├── pkg/
│   ├── models/           # 数据模型
│   ├── config/           # 配置
│   └── utils/            # 工具
├── configs/
│   └── config.yaml
├── go.mod
└── Makefile
```

## 3. 开发命令

### 3.1 前端命令

```bash
# 开发
npm run dev

# 构建
npm run build

# 预览构建
npm run preview

# 代码检查
npm run lint

# 类型检查
npm run type-check

# 测试
npm run test
```

### 3.2 后端命令

```bash
# 开发
go run cmd/server/main.go

# 构建
go build -o server cmd/server/main.go

# 测试
go test ./...

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 运行迁移
go run cmd/migrate/main.go
```

## 4. 代码规范

### 4.1 TypeScript 规范

```typescript
// ✅ 好的命名
interface UserResponse {
  code: number;
  message: string;
  data: UserData;
}

interface UserData {
  userId: string;
  userName: string;
}

// ❌ 避免的命名
interface res {
  c: number;
  m: string;
}
```

### 4.2 Go 规范

```go
// ✅ 好的命名
type VisitRequest struct {
    StudentID   string   `json:"student_id"`
    Symptoms    []Symptom `json:"symptoms"`
    Temperature float32  `json:"temperature"`
}

// ❌ 避免的命名
type req struct {
    id string
    s  []string
}
```

### 4.3 注释规范

**TypeScript:**
```typescript
/**
 * 创建就诊记录
 * @param symptoms - 症状列表
 * @param temperature - 体温
 * @returns 就诊记录 ID
 */
async function createVisit(symptoms: Symptom[], temperature: number): Promise<string> {
  // ...
}
```

**Go:**
```go
// CreateVisit 创建就诊记录
// 参数:
//   - symptoms: 症状列表
//   - temperature: 体温
// 返回:
//   - 就诊记录 ID
//   - 错误信息
func CreateVisit(symptoms []Symptom, temperature float32) (string, error) {
    // ...
}
```

## 5. Git 工作流

### 5.1 分支策略

```
main          - 主分支（生产环境）
develop       - 开发分支
feature/*     - 功能分支
bugfix/*      - 修复分支
hotfix/*      - 紧急修复
```

### 5.2 开发流程

```bash
# 1. 从 develop 创建功能分支
git checkout develop
git pull
git checkout -b feature/visit-module

# 2. 开发并提交
git add .
git commit -m "feat: 创建就诊记录功能"

# 3. 推送到远程
git push origin feature/visit-module

# 4. 创建 Pull Request
# 在 GitHub/GitLab 上创建 PR，请求合并到 develop

# 5. Code Review
# 等待团队成员 Review

# 6. 合并
# Review 通过后合并到 develop
```

### 5.3 提交规范（Conventional Commits）

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型：**
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

**示例：**
```bash
feat(visits): 创建就诊记录功能

- 实现患者端填写病情
- 实现医生端接诊功能
- 添加药物相互作用检查

Closes #123
```

## 6. 调试技巧

### 6.1 前端调试

```typescript
// 使用 React DevTools
// 使用 Redux DevTools（如果用 Redux）

// 调试 API 请求
axios.interceptors.request.use(config => {
  console.log('Request:', config);
  return config;
});

axios.interceptors.response.use(response => {
  console.log('Response:', response);
  return response;
});
```

### 6.2 后端调试

```go
// 使用日志
log.Printf("Processing visit request: %+v", request)

// 使用调试器
// VS Code: 安装 Go 扩展，配置 launch.json
// Delve: dlv debug cmd/server/main.go
```

## 7. 常见问题

### Q1: 前端启动失败？
```bash
# 清理缓存
rm -rf node_modules package-lock.json
npm install
```

### Q2: 后端连接数据库失败？
```bash
# 检查 PostgreSQL 是否运行
pg_isready

# 检查配置文件
cat configs/config.yaml
```

### Q3: Git 冲突如何解决？
```bash
# 1. 拉取最新代码
git pull origin develop

# 2. 解决冲突文件
# 编辑冲突文件，保留需要的代码

# 3. 标记解决
git add <conflicted-file>

# 4. 完成合并
git commit
```

---

*最后更新：2026-02-23*
*维护者：小饶 (Qwen)*
