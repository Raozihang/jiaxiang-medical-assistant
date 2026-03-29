import { ReloadOutlined } from "@ant-design/icons";
import {
  Button,
  Card,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import { listSafetyAlerts, resolveSafetyAlert, type SafetyAlert } from "@/shared/api/safety";

type SafetyAlertRow = {
  id: string;
  level: string;
  type: string;
  title: string;
  description: string;
  source: string;
  status: string;
  createdAt: string;
  resolvedAt: string;
};

function formatDate(value?: string) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function toRow(item: SafetyAlert): SafetyAlertRow {
  return {
    id: item.id,
    level: item.level,
    type: item.type,
    title: item.title,
    description: item.description,
    source: item.source,
    status: item.status,
    createdAt: formatDate(item.created_at),
    resolvedAt: formatDate(item.resolved_at),
  };
}

function levelColor(level: string) {
  if (["critical", "high", "urgent"].includes(level)) {
    return "red";
  }
  if (["medium", "warning"].includes(level)) {
    return "orange";
  }
  return "blue";
}

function statusColor(status: string) {
  if (["resolved", "closed", "done"].includes(status)) {
    return "green";
  }
  if (["open", "new"].includes(status)) {
    return "red";
  }
  return "blue";
}

export function SafetyPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [rows, setRows] = useState<SafetyAlertRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [resolvingId, setResolvingId] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState<string>("open");

  const fetchAlerts = useCallback(
    async (targetPage: number, targetPageSize: number, targetStatus: string) => {
      setLoading(true);
      try {
        const data = await listSafetyAlerts({
          page: targetPage,
          pageSize: targetPageSize,
          status: targetStatus === "all" ? undefined : targetStatus,
        });
        setRows(data.items.map(toRow));
        setPage(data.page || targetPage);
        setPageSize(data.page_size || targetPageSize);
        setTotal(data.total);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "告警列表获取失败"));
      } finally {
        setLoading(false);
      }
    },
    [messageApi],
  );

  useEffect(() => {
    void fetchAlerts(page, pageSize, statusFilter);
  }, [fetchAlerts, page, pageSize, statusFilter]);

  const handleResolve = async (id: string) => {
    setResolvingId(id);
    try {
      await resolveSafetyAlert(id);
      messageApi.success("告警已标记为已处理");
      await fetchAlerts(page, pageSize, statusFilter);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "告警处理失败"));
    } finally {
      setResolvingId(null);
    }
  };

  const columns: ColumnsType<SafetyAlertRow> = [
    { title: "告警 ID", dataIndex: "id", width: 180 },
    {
      title: "级别",
      dataIndex: "level",
      width: 120,
      render: (value: string) => <Tag color={levelColor(value)}>{value}</Tag>,
    },
    { title: "类型", dataIndex: "type", width: 120 },
    { title: "标题", dataIndex: "title", width: 180 },
    { title: "描述", dataIndex: "description" },
    { title: "来源", dataIndex: "source", width: 120 },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (value: string) => <Tag color={statusColor(value)}>{value}</Tag>,
    },
    { title: "触发时间", dataIndex: "createdAt", width: 180 },
    { title: "处理时间", dataIndex: "resolvedAt", width: 180 },
    {
      title: "操作",
      width: 140,
      render: (_, row) => {
        const resolved = ["resolved", "closed", "done"].includes(row.status);
        if (resolved) {
          return <Typography.Text type="secondary">已处理</Typography.Text>;
        }

        return (
          <Popconfirm
            title="确认标记为已处理？"
            onConfirm={() => void handleResolve(row.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" loading={resolvingId === row.id}>
              标记已处理
            </Button>
          </Popconfirm>
        );
      },
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        安全告警
      </Typography.Title>

      <Card
        extra={
          <Space>
            <Select
              value={statusFilter}
              style={{ width: 180 }}
              options={[
                { value: "open", label: "仅未处理" },
                { value: "resolved", label: "仅已处理" },
                { value: "all", label: "全部" },
              ]}
              onChange={(value) => {
                setStatusFilter(value);
                setPage(1);
              }}
            />
            <Button icon={<ReloadOutlined />} onClick={() => void fetchAlerts(page, pageSize, statusFilter)}>
              刷新
            </Button>
          </Space>
        }
      >
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

