# 05｜API 规范

## 1. 基础约定
- Base URL：`/api/v1`
- Content-Type：`application/json`
- 认证方式：`Authorization: Bearer <token>`
- 时间格式：ISO 8601（UTC）

## 2. 统一响应结构
```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "req_xxx",
  "timestamp": "2026-02-23T12:00:00Z"
}
```

- `code = 0` 表示成功；
- 非 0 表示业务错误；
- HTTP 状态码表达传输层语义（400/401/403/404/500）。

## 3. 分页规范
请求参数：`page`、`page_size`（默认 1/20，最大 100）

响应示例：
```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 120
}
```

## 4. 主要接口（P0）
| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 健康检查 |
| POST | `/auth/login` | 登录（占位，后接 OAuth） |
| GET | `/visits` | 就诊记录列表 |
| POST | `/visits` | 创建就诊记录 |
| GET | `/visits/:id` | 就诊详情 |
| PATCH | `/visits/:id` | 更新诊断/医嘱/去向 |
| GET | `/medicines` | 药品列表 |
| POST | `/medicines/inbound` | 药品入库 |
| POST | `/medicines/outbound` | 药品出库 |
| GET | `/reports/overview` | 管理看板基础数据 |

## 5. 错误码建议
| code | 含义 |
|---|---|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未认证 |
| 1003 | 无权限 |
| 2001 | 资源不存在 |
| 3001 | 库存不足 |
| 5000 | 系统内部错误 |

## 6. 版本策略
- 当前版本：`v1`
- 破坏性变更：新增 `v2` 路径；
- 非破坏性变更：仅追加字段，不删除旧字段。
