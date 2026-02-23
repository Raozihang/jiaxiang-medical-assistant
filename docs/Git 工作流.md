# Git 工作流

## 1. 分支模型
- `main`：稳定分支，仅合并已通过验证的版本；
- `develop`：日常集成分支；
- `feature/*`：新功能开发；
- `bugfix/*`：常规缺陷修复；
- `hotfix/*`：线上紧急修复。

## 2. 标准开发流程
1. 从 `develop` 拉出分支；
2. 小步提交并保持可构建；
3. 推送远端并发起 PR；
4. 通过评审和检查后合并回 `develop`；
5. 发布时由 `develop` 合并到 `main` 并打标签。

## 3. 提交规范（Conventional Commits）
```text
<type>(<scope>): <subject>
```

常用类型：
- `feat` 新功能
- `fix` 缺陷修复
- `docs` 文档变更
- `refactor` 重构
- `test` 测试
- `chore` 构建或工具链

示例：
```bash
git commit -m "feat(visit): 新增就诊详情接口"
git commit -m "fix(medicine): 修复库存扣减边界问题"
```

## 4. PR 要求
- 标题清晰描述变更目标；
- 描述中包含：背景、方案、影响范围、验证方式；
- 涉及 API/数据模型变更时同步更新 `docs/`；
- 至少一位同学 Review 通过后再合并。

## 5. 冲突处理
- 功能分支每天至少同步一次 `develop`；
- 冲突优先保留最新业务规则，避免机械覆盖；
- 解决冲突后重新执行项目自检命令。

## 6. 发布建议
- 使用语义化版本：`MAJOR.MINOR.PATCH`；
- 发布前冻结变更窗口并完成回归；
- 发布后记录变更日志与回滚预案。
