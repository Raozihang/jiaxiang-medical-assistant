# API 设计规范

## 1. 设计原则

- **RESTful 风格** - 资源导向，使用 HTTP 方法
- **统一响应格式** - 所有接口返回相同结构
- **版本控制** - URL 前缀 `/api/v1/`
- **认证授权** - JWT Token
- **错误码规范** - 统一错误码体系

## 2. 统一响应格式

### 2.1 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 实际数据
  },
  "timestamp": 1708646400000
}
```

### 2.2 错误响应

```json
{
  "code": 1001,
  "message": "参数错误",
  "data": null,
  "timestamp": 1708646400000,
  "details": {
    "field": "student_id",
    "reason": "必填字段"
  }
}
```

### 2.3 Go 响应结构

```go
type Response struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
    Details   interface{} `json:"details,omitempty"`
}
```

## 3. 错误码规范

### 3.1 错误码分类

| 错误码范围 | 分类 | 说明 |
|-----------|------|------|
| 0 | 成功 | 请求成功 |
| 1000-1999 | 客户端错误 | 参数错误、认证失败等 |
| 2000-2999 | 服务端错误 | 数据库错误、系统异常等 |
| 3000-3999 | 业务错误 | 业务逻辑错误 |

### 3.2 常用错误码

| 错误码 | 说明 | HTTP 状态码 |
|-------|------|-----------|
| 0 | 成功 | 200 |
| 1001 | 参数错误 | 400 |
| 1002 | 认证失败 | 401 |
| 1003 | 权限不足 | 403 |
| 1004 | 资源不存在 | 404 |
| 2001 | 数据库错误 | 500 |
| 2002 | 系统异常 | 500 |
| 3001 | 药品库存不足 | 400 |
| 3002 | 药物相互作用 | 400 |

## 4. API 接口列表

### 4.1 身份认证

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/auth/login` | POST | 登录（OAuth） |
| `/api/v1/auth/logout` | POST | 登出 |
| `/api/v1/auth/refresh` | POST | 刷新 Token |
| `/api/v1/auth/me` | GET | 获取当前用户信息 |

### 4.2 患者端

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/patient/visits` | POST | 创建就诊记录 |
| `/api/v1/patient/visits/history` | GET | 查看历史就诊 |
| `/api/v1/patient/destination` | POST | 填写离开去向 |

### 4.3 医生端

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/doctor/visits` | GET | 获取就诊列表 |
| `/api/v1/doctor/visits/:id` | GET | 查看就诊详情 |
| `/api/v1/doctor/visits/:id/diagnosis` | POST | 提交诊断 |
| `/api/v1/doctor/visits/:id/prescription` | POST | 开具处方 |
| `/api/v1/doctor/visits/:id/destination` | POST | 填写学生流向 |
| `/api/v1/doctor/patients/:id/history` | GET | 查看患者历史病历 |

### 4.4 药品管理

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/medicines` | GET | 获取药品列表 |
| `/api/v1/medicines/:id` | GET | 获取药品详情 |
| `/api/v1/medicines/inbound` | POST | 药品入库 |
| `/api/v1/medicines/outbound` | POST | 药品出库 |
| `/api/v1/medicines/stock-alert` | GET | 库存预警列表 |
| `/api/v1/medicines/expiry-alert` | GET | 近效期预警列表 |

### 4.5 AI 服务

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/ai/analyze` | POST | 症状分析 |
| `/api/v1/ai/triage` | POST | 智能分诊 |
| `/api/v1/ai/recommend` | POST | 药品推荐 |
| `/api/v1/ai/interaction-check` | POST | 药物相互作用检查 |
| `/api/v1/ai/similar-cases` | GET | 相似病例推荐 |

### 4.6 报表系统

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/reports/daily` | GET | 日报 |
| `/api/v1/reports/weekly` | GET | 周报 |
| `/api/v1/reports/monthly` | GET | 月报 |
| `/api/v1/reports/epidemic-warning` | GET | 流行病预警 |

## 5. 请求示例

### 5.1 创建就诊记录

**请求：**
```http
POST /api/v1/patient/visits
Content-Type: application/json
Authorization: Bearer <token>

{
  "symptoms": [
    {
      "name": "发热",
      "severity": 2,
      "description": "体温 38.5°C"
    },
    {
      "name": "咳嗽",
      "severity": 1,
      "description": "干咳"
    }
  ],
  "temperature": 38.5,
  "chiefComplaint": "发热咳嗽 1 天"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "visitId": "550e8400-e29b-41d4-a716-446655440000",
    "studentId": "20240001",
    "createdAt": "2026-02-23T10:30:00Z"
  },
  "timestamp": 1708646400000
}
```

### 5.2 开具处方

**请求：**
```http
POST /api/v1/doctor/visits/:id/prescription
Content-Type: application/json
Authorization: Bearer <token>

{
  "medicines": [
    {
      "medicineId": "med-001",
      "dosage": "1 片",
      "frequency": "每日 3 次",
      "duration": 3
    }
  ],
  "advice": "多喝水，注意休息"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "prescriptionId": "pres-001",
    "interactionCheck": {
      "hasInteraction": false,
      "warnings": []
    }
  },
  "timestamp": 1708646400000
}
```

## 6. 认证方式

### 6.1 JWT Token

**请求头：**
```
Authorization: Bearer <token>
```

**Token 结构：**
```json
{
  "userId": "user-001",
  "role": "doctor",
  "exp": 1708732800,
  "iat": 1708646400
}
```

### 6.2 Token 有效期

| Token 类型 | 有效期 | 用途 |
|-----------|--------|------|
| Access Token | 2 小时 | API 请求认证 |
| Refresh Token | 7 天 | 刷新 Access Token |

## 7. 限流策略

| 接口类型 | 限流 | 说明 |
|---------|------|------|
| 普通接口 | 100 次/分钟 | 常规业务接口 |
| AI 接口 | 20 次/分钟 | AI 服务成本高 |
| 报表接口 | 10 次/分钟 | 大数据量查询 |

## 8. 版本控制

- **当前版本：** v1
- **URL 前缀：** `/api/v1/`
- **弃用策略：** 新版本发布后，旧版本保留 3 个月

---

*最后更新：2026-02-23*
*维护者：小饶 (Qwen)*
