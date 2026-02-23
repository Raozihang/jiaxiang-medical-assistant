# 贡献指南（Contributing）

感谢参与嘉祥智能医务室助手项目。请在提交代码前遵循以下约定。

## 1. 开发前准备
- 阅读 `PRD.md` 与 `docs/README.md`；
- 对照 `docs/开发规范.md`、`docs/Git 工作流.md`；
- 在本地完成环境初始化：
  - 前端：`cd frontend && npm install`
  - 后端：`cd backend && go mod tidy`

## 2. 分支与提交
- 分支命名：`feature/*`、`bugfix/*`、`hotfix/*`；
- 提交信息遵循 Conventional Commits：
  - `feat(scope): ...`
  - `fix(scope): ...`
  - `docs: ...`
  - `chore: ...`
- 禁止直接向 `main` 推送。

## 3. 提交前自检
### 前端
```bash
cd frontend
npm run build
npm run lint
```

### 后端
```bash
cd backend
go test ./...
go vet ./...
```

## 4. PR 要求
- 描述变更目标与影响范围；
- 说明测试方式与结果；
- 涉及接口/数据结构变更时同步更新 `docs/`。

## 5. 安全与合规
- 禁止提交密钥、账号、生产库连接信息；
- 涉及学生健康数据时，优先使用脱敏样例；
- AI 相关建议必须明确“医生审核后生效”。
