# Git 工作流指南

## 1. 分支策略

### 1.1 分支类型

| 分支 | 命名 | 说明 | 合并目标 |
|------|------|------|---------|
| 主分支 | `main` | 生产环境代码 | - |
| 开发分支 | `develop` | 开发环境代码 | main (发布时) |
| 功能分支 | `feature/*` | 新功能开发 | develop |
| 修复分支 | `bugfix/*` | Bug 修复 | develop |
| 紧急修复 | `hotfix/*` | 生产紧急修复 | main + develop |
| 发布分支 | `release/*` | 版本发布准备 | main + develop |

### 1.2 分支图示

```
main     o─────────┬───────────────o (v1.0.0)
                   \             /
develop     o───────o───────────o
             \     / \         /
feature       o───o   o───────o
```

## 2. 开发流程

### 2.1 功能开发流程

```bash
# 1. 从 develop 创建功能分支
git checkout develop
git pull origin develop
git checkout -b feature/visit-module

# 2. 开发并提交（小步提交）
git add .
git commit -m "feat: 创建就诊表单组件"

git add .
git commit -m "feat: 添加症状选择器"

# 3. 推送到远程
git push origin feature/visit-module

# 4. 创建 Pull Request
# 在 GitHub/GitLab 上创建 PR
# 标题：feat: 就诊模块
# 描述：实现患者端就诊功能
# 关联 Issue: Closes #123

# 5. Code Review
# 等待团队成员 Review
# 根据评论修改代码

# 6. 合并到 develop
# PR 通过后合并
```

### 2.2 Bug 修复流程

```bash
# 1. 从 develop 创建修复分支
git checkout develop
git checkout -b bugfix/login-error

# 2. 修复并提交
git add .
git commit -m "fix: 修复登录 token 验证问题"

# 3. 创建 PR 并合并
git push origin bugfix/login-error
# 创建 PR 到 develop
```

### 2.3 紧急修复流程

```bash
# 1. 从 main 创建紧急修复分支
git checkout main
git checkout -b hotfix/critical-bug

# 2. 修复并提交
git add .
git commit -m "hotfix: 修复生产环境严重 bug"

# 3. 合并到 main 和 develop
git checkout main
git merge hotfix/critical-bug
git tag v1.0.1

git checkout develop
git merge hotfix/critical-bug

# 4. 删除分支
git branch -d hotfix/critical-bug
```

## 3. 提交规范

### 3.1 Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 3.2 Type 类型说明

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | feat(visits): 创建就诊记录 |
| `fix` | 修复 bug | fix(auth): 修复登录问题 |
| `docs` | 文档更新 | docs: 更新 API 文档 |
| `style` | 代码格式 | style: 格式化代码 |
| `refactor` | 重构 | refactor: 重构错误处理 |
| `test` | 测试 | test: 添加单元测试 |
| `chore` | 构建/工具 | chore: 更新依赖 |

### 3.3 提交示例

```bash
# 好的提交
git commit -m "feat(visits): 创建就诊记录功能

- 实现患者端填写病情
- 实现医生端接诊功能
- 添加药物相互作用检查

Closes #123"

# 避免的提交
git commit -m "update"
git commit -m "fix bug"
git commit -m "wip"
```

## 4. Pull Request 规范

### 4.1 PR 标题

```
<type>(<scope>): <description>

示例：
feat(visits): 创建就诊模块
fix(auth): 修复登录 token 验证问题
docs: 更新 API 文档
```

### 4.2 PR 描述模板

```markdown
## 变更说明
<!-- 描述这个 PR 做了什么 -->

## 关联 Issue
<!-- 关联的 Issue 编号 -->
Closes #123

## 测试步骤
<!-- 如何测试这些变更 -->
1. 
2. 
3. 

## 截图
<!-- 如果有 UI 变更，添加截图 -->

## 检查清单
- [ ] 代码已通过 ESLint/Go vet
- [ ] 已添加单元测试
- [ ] 已更新文档
- [ ] 已在本地测试通过
```

### 4.3 Code Review 要点

**审查者检查：**
- [ ] 代码功能是否正确
- [ ] 代码是否清晰易读
- [ ] 是否有潜在 bug
- [ ] 是否有性能问题
- [ ] 是否有安全隐患
- [ ] 测试是否充分
- [ ] 文档是否更新

**被审查者响应：**
- 及时回复评论
- 根据评论修改代码
- 有疑问时讨论解决

## 5. 版本发布

### 5.1 版本号规范

遵循 [Semantic Versioning](https://semver.org/)

```
主版本号。次版本号。修订号
MAJOR.MINOR.PATCH

- MAJOR: 不兼容的 API 变更
- MINOR: 向后兼容的功能新增
- PATCH: 向后兼容的问题修复
```

### 5.2 发布流程

```bash
# 1. 创建发布分支
git checkout develop
git checkout -b release/v1.0.0

# 2. 版本号更新
# 更新 package.json, CHANGELOG.md 等

# 3. 测试验证
npm run test
go test ./...

# 4. 合并到 main
git checkout main
git merge release/v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"

# 5. 合并回 develop
git checkout develop
git merge release/v1.0.0

# 6. 删除发布分支
git branch -d release/v1.0.0

# 7. 推送标签
git push origin v1.0.0
```

## 6. 冲突解决

### 6.1 预防冲突

- 小步提交，频繁拉取
- 及时沟通，避免撞车
- 大功能拆分成小 PR

### 6.2 解决冲突

```bash
# 1. 拉取最新代码
git checkout develop
git pull origin develop

# 2. 切换回功能分支
git checkout feature/visit-module

# 3. 合并 develop
git merge develop

# 4. 解决冲突文件
# 编辑冲突文件，保留需要的代码
# <<<<<<< HEAD
# 你的代码
# =======
# 别人的代码
# >>>>>>>

# 5. 标记解决
git add <conflicted-file>

# 6. 完成合并
git commit
```

## 7. 实用命令

```bash
# 查看分支
git branch
git branch -a  # 所有分支

# 查看提交历史
git log --oneline
git log --graph --oneline --all

# 撤销提交
git reset --soft HEAD~1  # 撤销提交，保留更改
git reset --hard HEAD~1  # 撤销提交，丢弃更改

# 暂存更改
git stash
git stash pop

# 查看差异
git diff
git diff HEAD~1

# 清理分支
git branch -d feature/xxx  # 删除本地分支
git push origin --delete feature/xxx  # 删除远程分支
```

---

*最后更新：2026-02-23*
*维护者：小饶 (Qwen)*
