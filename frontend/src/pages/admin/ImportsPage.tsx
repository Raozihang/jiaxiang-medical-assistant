import { Button, Card, Form, Input, Space, Table, Tag, Typography, message } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { createImportTask, listImportTasks, type ImportTask, type VisitImportItem } from "@/shared/api/imports";

type ImportForm = {
  visitsJson: string;
};

type ImportTaskRow = ImportTask & {
  progress: number;
  errorSummary: string;
};

const examplePayload = JSON.stringify(
  [
    {
      student_id: "20260002",
      symptoms: ["cough", "fever"],
      description: "Imported from history",
      diagnosis: "Common cold",
      prescription: ["Warm water"],
      destination: "observation",
    },
  ],
  null,
  2,
);

function statusColor(status: string) {
  switch (status) {
    case "completed":
      return "green";
    case "completed_with_errors":
      return "orange";
    case "failed":
      return "red";
    default:
      return "blue";
  }
}

function toRow(task: ImportTask): ImportTaskRow {
  const progress = task.total > 0 ? Math.round(((task.success + task.failed) / task.total) * 100) : 0;
  return {
    ...task,
    progress,
    errorSummary: task.errors.map((item) => `#${item.index + 1} ${item.message}`).join("; ") || "-",
  };
}

export function ImportsPage() {
  const [form] = Form.useForm<ImportForm>();
  const [rows, setRows] = useState<ImportTaskRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [messageApi, contextHolder] = message.useMessage();

  const fetchTasks = useCallback(async (targetPage: number, targetPageSize: number) => {
    setLoading(true);
    try {
      const data = await listImportTasks({ page: targetPage, pageSize: targetPageSize });
      setRows(data.items.map(toRow));
      setPage(data.page);
      setPageSize(data.page_size);
      setTotal(data.total);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchTasks(page, pageSize);
  }, [fetchTasks, page, pageSize]);

  const handleCreate = async (values: ImportForm) => {
    setSubmitting(true);
    try {
      const items = JSON.parse(values.visitsJson) as VisitImportItem[];
      if (!Array.isArray(items)) {
        throw new Error("JSON must be an array");
      }
      await createImportTask(items);
      messageApi.success("Import task created");
      await fetchTasks(1, pageSize);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "Failed to create import task");
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ColumnsType<ImportTaskRow> = [
    { title: "Task ID", dataIndex: "id", width: 220 },
    { title: "Status", dataIndex: "status", width: 180, render: (value: string) => <Tag color={statusColor(value)}>{value}</Tag> },
    { title: "Total", dataIndex: "total", width: 80 },
    { title: "Success", dataIndex: "success", width: 80 },
    { title: "Failed", dataIndex: "failed", width: 80 },
    { title: "Progress", dataIndex: "progress", width: 100, render: (value: number) => `${value}%` },
    { title: "Errors", dataIndex: "errorSummary" },
    { title: "Updated At", dataIndex: "updated_at", width: 180, render: (value: string) => new Date(value).toLocaleString() },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        Historical Visit Import
      </Typography.Title>
      <Card title="Create Import Task (JSON array)" extra={<Button onClick={() => void fetchTasks(page, pageSize)}>Refresh</Button>}>
        <Form layout="vertical" form={form} onFinish={(values) => void handleCreate(values)} initialValues={{ visitsJson: examplePayload }}>
          <Form.Item label="Import Data" name="visitsJson" rules={[{ required: true, message: "Please input import JSON" }]}>
            <Input.TextArea rows={10} placeholder={examplePayload} />
          </Form.Item>
          <Button htmlType="submit" type="primary" loading={submitting}>
            Submit Import
          </Button>
        </Form>
      </Card>
      <Card title="Import Task Status">
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
          scroll={{ x: 1200 }}
        />
      </Card>
    </Space>
  );
}
