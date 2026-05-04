import { ReloadOutlined } from "@ant-design/icons";
import { Button, Card, Form, Input, message, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  createImportTask,
  type ImportTask,
  listImportTasks,
  type VisitImportItem,
} from "@/shared/api/imports";
import { getStatusLabel } from "@/shared/labels/localization";

type ImportTaskRow = {
  id: string;
  status: string;
  total: number;
  success: number;
  failed: number;
  progress: number;
  errorSummary: string;
  createdAt: string;
  updatedAt: string;
};

type ImportForm = {
  visitsJson: string;
};

const examplePayload = `[
  {
    "student_id": "20260001",
    "symptoms": ["发热", "咳嗽"],
    "description": "纸质病历补录：学生晨检低热并伴咳嗽。",
    "diagnosis": "上呼吸道感染",
    "prescription": ["对乙酰氨基酚片"],
    "destination": "留观",
    "created_at": "2026-02-20T08:00:00Z"
  }
]`;

function formatDate(value: string) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function statusColor(status: string) {
  if (["completed", "success"].includes(status)) {
    return "green";
  }
  if (["failed"].includes(status)) {
    return "red";
  }
  if (["processing", "running"].includes(status)) {
    return "blue";
  }
  if (["completed_with_errors"].includes(status)) {
    return "orange";
  }
  return "default";
}

function toTaskRow(task: ImportTask): ImportTaskRow {
  const errorSummary =
    task.errors.length > 0
      ? task.errors.map((item) => `#${item.index}: ${item.message}`).join("; ")
      : "-";

  return {
    id: task.id,
    status: task.status,
    total: task.total,
    success: task.success,
    failed: task.failed,
    progress: task.progress,
    errorSummary,
    createdAt: formatDate(task.created_at),
    updatedAt: formatDate(task.updated_at),
  };
}

export function ImportsPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [form] = Form.useForm<ImportForm>();
  const [rows, setRows] = useState<ImportTaskRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);

  const fetchTasks = useCallback(
    async (targetPage: number, targetPageSize: number) => {
      setLoading(true);
      try {
        const data = await listImportTasks({ page: targetPage, pageSize: targetPageSize });
        setRows(data.items.map(toTaskRow));
        setPage(data.page || targetPage);
        setPageSize(data.page_size || targetPageSize);
        setTotal(data.total);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "导入任务获取失败"));
      } finally {
        setLoading(false);
      }
    },
    [messageApi],
  );

  useEffect(() => {
    void fetchTasks(page, pageSize);
  }, [fetchTasks, page, pageSize]);

  const handleCreate = async (values: ImportForm) => {
    let items: VisitImportItem[] = [];
    try {
      const parsed = JSON.parse(values.visitsJson);
      if (!Array.isArray(parsed)) {
        messageApi.error("导入数据必须是 JSON 数组");
        return;
      }
      items = parsed as VisitImportItem[];
    } catch {
      messageApi.error("JSON 格式不正确");
      return;
    }

    setSubmitting(true);
    try {
      await createImportTask(items);
      messageApi.success("导入任务已创建");
      await fetchTasks(1, pageSize);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "创建导入任务失败"));
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ColumnsType<ImportTaskRow> = [
    { title: "任务 ID", dataIndex: "id", width: 220 },
    {
      title: "状态",
      dataIndex: "status",
      width: 180,
      render: (value: string) => <Tag color={statusColor(value)}>{getStatusLabel(value)}</Tag>,
    },
    { title: "总数", dataIndex: "total", width: 80 },
    { title: "成功", dataIndex: "success", width: 80 },
    { title: "失败", dataIndex: "failed", width: 80 },
    { title: "进度", dataIndex: "progress", width: 100, render: (value) => `${value}%` },
    { title: "错误", dataIndex: "errorSummary", width: 320 },
    { title: "创建时间", dataIndex: "createdAt", width: 180 },
    { title: "更新时间", dataIndex: "updatedAt", width: 180 },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        历史数据导入
      </Typography.Title>

      <Card
        title="创建导入任务（JSON 数组）"
        extra={
          <Button icon={<ReloadOutlined />} onClick={() => void fetchTasks(page, pageSize)}>
            刷新列表
          </Button>
        }
      >
        <Form
          layout="vertical"
          form={form}
          onFinish={(values) => void handleCreate(values)}
          initialValues={{ visitsJson: examplePayload }}
        >
          <Form.Item
            label="导入数据"
            name="visitsJson"
            rules={[{ required: true, message: "请输入导入数据 JSON" }]}
          >
            <Input.TextArea rows={10} placeholder={examplePayload} />
          </Form.Item>
          <Button htmlType="submit" type="primary" loading={submitting}>
            提交导入
          </Button>
        </Form>
      </Card>

      <Card title="导入任务状态">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={rows}
          loading={loading}
          pagination={
            {
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              onChange: (targetPage: number, targetPageSize: number) => {
                setPage(targetPage);
                setPageSize(targetPageSize);
              },
            } as TablePaginationConfig
          }
          scroll={{ x: 1400 }}
        />
      </Card>
    </Space>
  );
}
