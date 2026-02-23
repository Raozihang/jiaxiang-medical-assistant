# 数据库设计文档

## 1. 数据库选型

- **主数据库：** PostgreSQL 15+
- **向量检索：** pgvector 0.5+（AI 知识库）
- **缓存：** Redis 7+（会话、热点数据）

## 2. 核心数据表

### 2.1 学生表 (students)

```sql
CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id VARCHAR(50) UNIQUE NOT NULL,  -- 学号
    name VARCHAR(100) NOT NULL,              -- 姓名
    class_id UUID,                            -- 班级 ID
    grade VARCHAR(20),                        -- 年级
    gender VARCHAR(10),                       -- 性别
    birthday DATE,                            -- 出生日期
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_students_class ON students(class_id);
CREATE INDEX idx_students_student_id ON students(student_id);
```

### 2.2 就诊记录表 (visits)

```sql
CREATE TABLE visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID REFERENCES students(id),
    doctor_id UUID,                           -- 医生 ID
    visit_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    symptoms JSONB,                           -- 症状列表
    temperature DECIMAL(4,1),                 -- 体温
    chief_complaint TEXT,                     -- 主诉
    diagnosis TEXT,                           -- 诊断结果
    prescription JSONB,                       -- 处方
    advice TEXT,                              -- 医嘱
    destination VARCHAR(50),                  -- 离开去向
    status VARCHAR(20) DEFAULT 'completed',   -- 状态
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_visits_student ON visits(student_id);
CREATE INDEX idx_visits_doctor ON visits(doctor_id);
CREATE INDEX idx_visits_time ON visits(visit_time);
CREATE INDEX idx_visits_symptoms ON visits USING GIN(symptoms);
```

### 2.3 药品表 (medicines)

```sql
CREATE TABLE medicines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,               -- 药品名称
    specification VARCHAR(100),               -- 规格
    manufacturer VARCHAR(200),                -- 生产厂家
    stock INTEGER DEFAULT 0,                  -- 库存数量
    safety_stock INTEGER DEFAULT 10,          -- 安全库存
    expiry_date DATE,                         -- 有效期
    batch_number VARCHAR(50),                 -- 批号
    price DECIMAL(10,2),                      -- 单价
    warnings JSONB,                           -- 禁忌症/不良反应
    interactions JSONB,                       -- 药物相互作用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_medicines_name ON medicines(name);
CREATE INDEX idx_medicines_expiry ON medicines(expiry_date);
CREATE INDEX idx_medicines_stock ON medicines(stock);
```

### 2.4 药品库存记录表 (medicine_stock_records)

```sql
CREATE TABLE medicine_stock_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    medicine_id UUID REFERENCES medicines(id),
    type VARCHAR(20) NOT NULL,                -- 类型：inbound/outbound
    quantity INTEGER NOT NULL,                -- 数量
    balance INTEGER,                          -- 结余
    operator_id UUID,                         -- 操作人 ID
    remark TEXT,                              -- 备注
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stock_records_medicine ON medicine_stock_records(medicine_id);
CREATE INDEX idx_stock_records_type ON medicine_stock_records(type);
```

### 2.5 医生表 (doctors)

```sql
CREATE TABLE doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    title VARCHAR(50),                        -- 职称
    phone VARCHAR(20),
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2.6 班级表 (classes)

```sql
CREATE TABLE classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,               -- 班级名称
    grade VARCHAR(20),                        -- 年级
    teacher_name VARCHAR(100),                -- 班主任
    teacher_phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2.7 AI 知识库表 (knowledge_base)

```sql
-- 需要 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE knowledge_base (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    content TEXT,
    category VARCHAR(50),                     -- 类别：medicine/disease/guideline
    embedding vector(768),                    -- 向量嵌入（768 维）
    metadata JSONB,                           -- 元数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_knowledge_embedding ON knowledge_base USING ivfflat(embedding vector_cosine_ops);
CREATE INDEX idx_knowledge_category ON knowledge_base(category);
```

## 3. 数据字典

### 3.1 就诊状态 (visit_status)

| 状态 | 说明 |
|------|------|
| pending | 待接诊 |
| in_progress | 接诊中 |
| completed | 已完成 |
| referred | 已转诊 |

### 3.2 离开去向 (destination)

| 去向 | 说明 |
|------|------|
| back_to_class | 返回教室 |
| back_to_dorm | 返回寝室 |
| leave_school | 离校 |
| referred | 转诊医院 |
| observation | 留观 |

### 3.3 症状严重程度 (severity)

| 等级 | 说明 |
|------|------|
| 0 | 无 |
| 1 | 轻度 |
| 2 | 中度 |
| 3 | 重度 |

## 4. 数据迁移

### 4.1 历史病历导入

```sql
-- 临时表
CREATE TABLE temp_visits (
    student_id VARCHAR(50),
    visit_time TIMESTAMP,
    symptoms TEXT,
    diagnosis TEXT,
    prescription TEXT
);

-- 导入后转换
INSERT INTO visits (student_id, visit_time, symptoms, diagnosis, prescription)
SELECT s.id, t.visit_time, t.symptoms::jsonb, t.diagnosis, t.prescription::jsonb
FROM temp_visits t
JOIN students s ON t.student_id = s.student_id;
```

### 4.2 药品数据导入

```sql
-- 从 Excel 导入药品数据
COPY medicines (name, specification, manufacturer, stock, expiry_date)
FROM '/path/to/medicines.csv'
DELIMITER ','
CSV HEADER;
```

## 5. 数据备份

### 5.1 定期备份

```bash
# 每日备份
pg_dump -U postgres medical_assistant > backup_$(date +%Y%m%d).sql

# 保留 7 天
find /backup -name "backup_*.sql" -mtime +7 -delete
```

### 5.2 恢复数据

```bash
psql -U postgres medical_assistant < backup_20260223.sql
```

## 6. 性能优化

### 6.1 索引策略

- 外键字段必建索引
- 查询条件字段建索引
- 时间范围查询用时间索引
- JSONB 字段用 GIN 索引

### 6.2 查询优化

```sql
-- 使用 EXPLAIN 分析查询计划
EXPLAIN ANALYZE
SELECT * FROM visits
WHERE student_id = 'xxx'
ORDER BY visit_time DESC
LIMIT 10;
```

### 6.3 分区表（可选）

```sql
-- 按月份分区就诊记录
CREATE TABLE visits_2026_02 PARTITION OF visits
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
```

---

*最后更新：2026-02-23*
*维护者：小饶 (Qwen)*
