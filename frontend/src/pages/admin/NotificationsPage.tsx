import { ReloadOutlined } from "@ant-design/icons";
import {
  Button,
  Card,
  Form,
  Input,
  message,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import {
  dispatchScenarioNotification,
  listNotificationLogs,
  sendNotification,
  type NotificationLog,
  type NotificationChannel,
  type NotificationScenario,
  type DispatchScenarioNotificationPayload,
  type SendNotificationPayload,
} from "@/shared/api/notifications";
import { getErrorMessage } from "@/shared/api/helpers";

type NotificationLogRow = {
  id: string;
  channel: string;
  status: string;
  receiver: string;
  message: string;
  sentAt: string;
  error: string;
};

type NotificationForm = {
  channel: NotificationChannel;
  receiver: string;
  message: string;
};

type ScenarioNotificationForm = {
  scenario: NotificationScenario;
  channel: NotificationChannel;
  receiver: string;
  student_name?: string;
  destination?: string;
  follow_up_at?: string;
  note?: string;
};

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

function toLogRow(item: NotificationLog): NotificationLogRow {
  return {
    id: item.id,
    channel: item.channel,
    status: item.status,
    receiver: item.receiver || "-",
    message: item.message,
    sentAt: formatDate(item.sent_at),
    error: item.error || "-",
  };
}

function statusColor(status: string) {
  if (["success", "sent", "ok"].includes(status)) {
    return "green";
  }
  if (["failed", "error"].includes(status)) {
    return "red";
  }
  if (["pending", "queued", "running"].includes(status)) {
    return "blue";
  }
  return "default";
}

export function NotificationsPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [manualForm] = Form.useForm<NotificationForm>();
  const [scenarioForm] = Form.useForm<ScenarioNotificationForm>();
  const [rows, setRows] = useState<NotificationLogRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [manualSubmitting, setManualSubmitting] = useState(false);
  const [scenarioSubmitting, setScenarioSubmitting] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);

  const fetchLogs = useCallback(
    async (targetPage: number, targetPageSize: number) => {
      setLoading(true);
      try {
        const data = await listNotificationLogs({ page: targetPage, pageSize: targetPageSize });
        setRows(data.items.map(toLogRow));
        setPage(data.page || targetPage);
        setPageSize(data.page_size || targetPageSize);
        setTotal(data.total);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "消息日志获取失败"));
      } finally {
        setLoading(false);
      }
    },
    [messageApi],
  );

  useEffect(() => {
    void fetchLogs(page, pageSize);
  }, [fetchLogs, page, pageSize]);

  const notifyAndRefresh = async (
    action: () => Promise<unknown>,
    successText: string,
    failureText: string,
    onSuccess?: () => void,
  ) => {
    try {
      await action();
      messageApi.success(successText);
      onSuccess?.();
      await fetchLogs(1, pageSize);
    } catch (error) {
      messageApi.error(getErrorMessage(error, failureText));
    }
  };

  const handleSend = async (values: NotificationForm) => {
    const payload: SendNotificationPayload = {
      channel: values.channel,
      receiver: values.receiver,
      message: values.message,
    };

    setManualSubmitting(true);
    try {
      await notifyAndRefresh(
        () => sendNotification(payload),
        "消息发送成功",
        "消息发送失败",
        () => manualForm.resetFields(),
      );
    } finally {
      setManualSubmitting(false);
    }
  };

  const handleScenarioDispatch = async (values: ScenarioNotificationForm) => {
    const payload: DispatchScenarioNotificationPayload = {
      scenario: values.scenario,
      channel: values.channel,
      receiver: values.receiver,
      student_name: values.student_name?.trim() || undefined,
      destination: values.destination?.trim() || undefined,
      follow_up_at: values.follow_up_at?.trim() || undefined,
      note: values.note?.trim() || undefined,
    };

    setScenarioSubmitting(true);
    try {
      await notifyAndRefresh(
        () => dispatchScenarioNotification(payload),
        "场景化推送发送成功",
        "场景化推送发送失败",
        () => scenarioForm.resetFields(),
      );
    } finally {
      setScenarioSubmitting(false);
    }
  };

  const columns: ColumnsType<NotificationLogRow> = [
    { title: "日志 ID", dataIndex: "id", width: 200 },
    {
      title: "通道",
      dataIndex: "channel",
      width: 120,
      render: (value: string) => <Tag>{value}</Tag>,
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (value: string) => <Tag color={statusColor(value)}>{value}</Tag>,
    },
    { title: "接收人", dataIndex: "receiver", width: 220 },
    { title: "消息内容", dataIndex: "message" },
    { title: "发送时间", dataIndex: "sentAt", width: 180 },
    { title: "错误", dataIndex: "error", width: 220 },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        消息通知中心
      </Typography.Title>

      <Card
        title="发送通知（WeChat / DingTalk）"
        extra={
          <Button icon={<ReloadOutlined />} onClick={() => void fetchLogs(page, pageSize)}>
            刷新日志
          </Button>
        }
      >
        <Form layout="vertical" form={manualForm} onFinish={(values) => void handleSend(values)}>
          <Space align="start" wrap>
            <Form.Item
              label="发送通道"
              name="channel"
              rules={[{ required: true, message: "请选择发送通道" }]}
            >
              <Select
                style={{ width: 160 }}
                options={[
                  { value: "wechat", label: "WeChat" },
                  { value: "dingtalk", label: "DingTalk" },
                ]}
              />
            </Form.Item>

            <Form.Item
              label="通知内容"
              name="message"
              rules={[{ required: true, message: "请输入通知内容" }]}
            >
              <Input.TextArea rows={2} style={{ width: 380 }} placeholder="请输入消息内容" />
            </Form.Item>

            <Form.Item
              label="接收人"
              name="receiver"
              rules={[{ required: true, message: "请输入接收人标识" }]}
            >
              <Input style={{ width: 240 }} placeholder="如：parent_group_1" />
            </Form.Item>

            <Form.Item label=" " style={{ marginBottom: 0 }}>
              <Button htmlType="submit" type="primary" loading={manualSubmitting}>
                发送
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      <Card title="场景化推送">
        <Form
          layout="vertical"
          form={scenarioForm}
          onFinish={(values) => void handleScenarioDispatch(values)}
        >
          <Space align="start" wrap>
            <Form.Item
              label="推送场景"
              name="scenario"
              rules={[{ required: true, message: "请选择推送场景" }]}
            >
              <Select
                style={{ width: 220 }}
                options={[
                  { value: "visit_completed", label: "visit_completed（到访完成）" },
                  { value: "observation_notice", label: "observation_notice（观察通知）" },
                  { value: "follow_up_reminder", label: "follow_up_reminder（复诊提醒）" },
                ]}
              />
            </Form.Item>

            <Form.Item
              label="发送通道"
              name="channel"
              rules={[{ required: true, message: "请选择发送通道" }]}
            >
              <Select
                style={{ width: 160 }}
                options={[
                  { value: "wechat", label: "WeChat" },
                  { value: "dingtalk", label: "DingTalk" },
                ]}
              />
            </Form.Item>

            <Form.Item
              label="接收人"
              name="receiver"
              rules={[{ required: true, message: "请输入接收人标识" }]}
            >
              <Input style={{ width: 220 }} placeholder="如：parent_group_1" />
            </Form.Item>

            <Form.Item label="学生姓名" name="student_name">
              <Input style={{ width: 180 }} placeholder="可选" />
            </Form.Item>

            <Form.Item label="目的地" name="destination">
              <Input style={{ width: 180 }} placeholder="可选" />
            </Form.Item>

            <Form.Item label="复诊时间" name="follow_up_at">
              <Input style={{ width: 220 }} placeholder="可选，如 2026-02-24 09:30" />
            </Form.Item>

            <Form.Item label="备注" name="note">
              <Input.TextArea rows={2} style={{ width: 260 }} placeholder="可选" />
            </Form.Item>

            <Form.Item label=" " style={{ marginBottom: 0 }}>
              <Button htmlType="submit" type="primary" loading={scenarioSubmitting}>
                发送场景化推送
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      <Card title="发送日志">
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
          scroll={{ x: 1300 }}
        />
      </Card>
    </Space>
  );
}

