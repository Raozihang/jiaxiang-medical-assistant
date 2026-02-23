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
  listNotificationLogs,
  sendNotification,
  type NotificationLog,
  type NotificationChannel,
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
  const [form] = Form.useForm<NotificationForm>();
  const [rows, setRows] = useState<NotificationLogRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
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

  const handleSend = async (values: NotificationForm) => {
    const payload: SendNotificationPayload = {
      channel: values.channel,
      receiver: values.receiver,
      message: values.message,
    };

    setSubmitting(true);
    try {
      await sendNotification(payload);
      messageApi.success("消息发送成功");
      form.resetFields();
      await fetchLogs(1, pageSize);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "消息发送失败"));
    } finally {
      setSubmitting(false);
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
        <Form layout="vertical" form={form} onFinish={(values) => void handleSend(values)}>
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
              <Button htmlType="submit" type="primary" loading={submitting}>
                发送
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

