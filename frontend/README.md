# 前端工程（React + TypeScript）

## 技术栈
- React 18+
- TypeScript
- Vite
- Ant Design
- React Router
- Axios
- Zustand

## 启动
```bash
npm install
npm run dev
```

## 构建
```bash
npm run build
```

## 目录
```text
src/
├─app/        # 路由与应用装配
├─pages/      # 页面层
├─shared/     # 通用能力（api/config/layout）
└─styles/     # 全局样式
```

## 环境变量
复制 `.env.example` 为 `.env.local` 后按需修改：
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=智慧医务室系统
```

## Lint
`ash
npm run lint
`

