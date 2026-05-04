import { ReloadOutlined, WechatOutlined } from "@ant-design/icons";
import { Button, Card, message, Popconfirm, Select, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  dispatchScenarioNotification,
  listStudentContacts,
} from "@/shared/api/notifications";
import { listSafetyAlerts, resolveSafetyAlert, type SafetyAlert } from "@/shared/api/safety";
import {
  getSafetyLevelLabel,
  getSafetyTypeLabel,
  getStatusLabel,
} from "@/shared/labels/localization";

type SafetyAlertRow = {
  id: string;
  level: string;
  type: string;
  title: string;
  description: string;
  source: string;
  studentId: string;
  studentName?: string;
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
    studentId: item.student_id,
    studentName: item.student_name,
    status: item.status,
    createdAt: formatDate(item.created_at),
    resolvedAt: formatDate(item.resolved_at),
  };
}

function formatStudent(row: SafetyAlertRow) {
  const studentName = row.studentName?.trim();
  const studentId = (row.studentId || row.source || "").trim();
  if (studentName && studentId && studentName !== studentId) {
    return `${studentName}（${studentId}）`;
  }
  return studentName || studentId || "该学生";
}

function fallbackWechatReceiver(row: SafetyAlertRow) {
  const studentId = (row.studentId || row.source || "unknown").trim();
  return `class_teacher_parent_${studentId}`;
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
  const [sendingWechatId, setSendingWechatId] = useState<string | null>(null);

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

  const handleSendObservationWechat = async (row: SafetyAlertRow) => {
    setSendingWechatId(row.id);
    try {
      const studentId = (row.studentId || row.source).trim();
      const contacts = studentId
        ? await listStudentContacts({ page: 1, pageSize: 1, keyword: studentId })
        : null;
      const contact = contacts?.items.find((item) => item.student_id === studentId) ?? contacts?.items[0];
      const receiver = contact?.guardian_phone?.trim() || fallbackWechatReceiver(row);

      await dispatchScenarioNotification({
        scenario: "observation_notice",
        channel: "wechat",
        receiver,
        student_name: contact?.student_name?.trim() || row.studentName || row.studentId,
        destination: "医务室留观区",
        note: `${row.description || "留观超时"}，请班主任或家长及时关注。`,
      });

      messageApi.success(`已发送 ${formatStudent(row)} 的留观微信通知`);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "微信通知发送失败"));
    } finally {
      setSendingWechatId(null);
    }
  };

  const columns: ColumnsType<SafetyAlertRow> = [
    { title: "告警 ID", dataIndex: "id", width: 180 },
    {
      title: "级别",
      dataIndex: "level",
      width: 120,
      render: (value: string) => <Tag color={levelColor(value)}>{getSafetyLevelLabel(value)}</Tag>,
    },
    {
      title: "类型",
      dataIndex: "type",
      width: 140,
      render: (value: string) => getSafetyTypeLabel(value),
    },
    { title: "标题", dataIndex: "title", width: 180 },
    { title: "描述", dataIndex: "description" },
    {
      title: "来源",
      dataIndex: "source",
      width: 120,
      render: (value: string) => (value === "system" ? "系统" : value),
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (value: string) => <Tag color={statusColor(value)}>{getStatusLabel(value)}</Tag>,
    },
    { title: "触发时间", dataIndex: "createdAt", width: 180 },
    { title: "处理时间", dataIndex: "resolvedAt", width: 180 },
    {
      title: "操作",
      width: 260,
      render: (_, row) => {
        const resolved = ["resolved", "closed", "done"].includes(row.status);
        if (resolved) {
          return <Typography.Text type="secondary">已处理</Typography.Text>;
        }

        return (
          <Space size={8} wrap>
            {row.type === "observation_timeout" ? (
              <Button
                icon={<WechatOutlined />}
                loading={sendingWechatId === row.id}
                onClick={() => void handleSendObservationWechat(row)}
              >
                发微信
              </Button>
            ) : null}
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
          </Space>
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
            <Button
              icon={<ReloadOutlined />}
              onClick={() => void fetchAlerts(page, pageSize, statusFilter)}
            >
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
