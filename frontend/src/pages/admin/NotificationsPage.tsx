import { ReloadOutlined } from "@ant-design/icons";
import { Button, Card, Form, Input, message, Select, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  type DispatchScenarioNotificationPayload,
  dispatchScenarioNotification,
  listNotificationLogs,
  listOutboundCalls,
  listStudentContacts,
  type NotificationChannel,
  type NotificationLog,
  type NotificationScenario,
  type OutboundCall,
  retryOutboundCall,
  type SendNotificationPayload,
  type StudentContact,
  sendNotification,
  updateStudentContact,
} from "@/shared/api/notifications";
import {
  getChannelLabel,
  getProviderLabel,
  getStatusLabel,
  getTriggerSourceLabel,
} from "@/shared/labels/localization";

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

type ContactForm = {
  student_id: string;
  student_name?: string;
  guardian_name?: string;
  guardian_phone?: string;
  guardian_relation?: string;
};

const channelOptions = [
  { value: "wechat", label: "微信" },
  { value: "dingtalk", label: "钉钉" },
];

const scenarioOptions = [
  { value: "visit_completed", label: "就诊完成" },
  { value: "observation_notice", label: "留观通知" },
  { value: "follow_up_reminder", label: "复诊提醒" },
];

const destinationOptions = [
  { value: "留观", label: "留观" },
  { value: "返回班级", label: "返回班级" },
  { value: "转诊", label: "转诊" },
  { value: "紧急处理", label: "紧急处理" },
  { value: "转外院", label: "转外院" },
  { value: "离校就医", label: "离校就医" },
  { value: "返回宿舍", label: "返回宿舍" },
  { value: "离校回家", label: "离校回家" },
];

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

function statusColor(status: string) {
  if (["success", "sent", "ok", "connected"].includes(status)) {
    return "green";
  }
  if (["failed", "error", "busy", "no_answer", "cancelled"].includes(status)) {
    return "red";
  }
  if (["pending", "queued", "running", "requested", "processing"].includes(status)) {
    return "blue";
  }
  return "default";
}

export function NotificationsPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [manualForm] = Form.useForm<NotificationForm>();
  const [scenarioForm] = Form.useForm<ScenarioNotificationForm>();
  const [contactForm] = Form.useForm<ContactForm>();

  const [manualSubmitting, setManualSubmitting] = useState(false);
  const [scenarioSubmitting, setScenarioSubmitting] = useState(false);
  const [contactSubmitting, setContactSubmitting] = useState(false);
  const [retryingId, setRetryingId] = useState<string | null>(null);

  const [logs, setLogs] = useState<NotificationLog[]>([]);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logPage, setLogPage] = useState(1);
  const [logPageSize, setLogPageSize] = useState(10);
  const [logTotal, setLogTotal] = useState(0);

  const [contacts, setContacts] = useState<StudentContact[]>([]);
  const [contactsLoading, setContactsLoading] = useState(false);
  const [contactKeyword, setContactKeyword] = useState("");
  const [contactPage, setContactPage] = useState(1);
  const [contactPageSize, setContactPageSize] = useState(10);
  const [contactTotal, setContactTotal] = useState(0);

  const [calls, setCalls] = useState<OutboundCall[]>([]);
  const [callsLoading, setCallsLoading] = useState(false);
  const [callKeyword, setCallKeyword] = useState("");
  const [callStatus, setCallStatus] = useState<string | undefined>(undefined);
  const [callPage, setCallPage] = useState(1);
  const [callPageSize, setCallPageSize] = useState(10);
  const [callTotal, setCallTotal] = useState(0);

  const fetchLogs = useCallback(
    async (page: number, pageSize: number) => {
      setLogsLoading(true);
      try {
        const data = await listNotificationLogs({ page, pageSize });
        setLogs(data.items);
        setLogPage(data.page || page);
        setLogPageSize(data.page_size || pageSize);
        setLogTotal(data.total);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "获取通知日志失败"));
      } finally {
        setLogsLoading(false);
      }
    },
    [messageApi],
  );

  const fetchContacts = useCallback(
    async (page: number, pageSize: number, keyword: string) => {
      setContactsLoading(true);
      try {
        const data = await listStudentContacts({ page, pageSize, keyword });
        setContacts(data.items);
        setContactPage(data.page || page);
        setContactPageSize(data.page_size || pageSize);
        setContactTotal(data.total);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "获取家长联系人失败"));
      } finally {
        setContactsLoading(false);
      }
    },
    [messageApi],
  );

  const fetchCalls = useCallback(
    async (page: number, pageSize: number, status?: string, keyword?: string) => {
      setCallsLoading(true);
      try {
        const data = await listOutboundCalls({ page, pageSize, status, keyword });
        setCalls(data.items);
        setCallPage(data.page || page);
        setCallPageSize(data.page_size || pageSize);
        setCallTotal(data.total);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "获取智能外呼失败"));
      } finally {
        setCallsLoading(false);
      }
    },
    [messageApi],
  );

  useEffect(() => {
    void fetchLogs(logPage, logPageSize);
  }, [fetchLogs, logPage, logPageSize]);

  useEffect(() => {
    void fetchContacts(contactPage, contactPageSize, contactKeyword);
  }, [contactKeyword, contactPage, contactPageSize, fetchContacts]);

  useEffect(() => {
    void fetchCalls(callPage, callPageSize, callStatus, callKeyword);
  }, [callKeyword, callPage, callPageSize, callStatus, fetchCalls]);

  const refreshAll = async () => {
    await Promise.all([
      fetchLogs(logPage, logPageSize),
      fetchContacts(contactPage, contactPageSize, contactKeyword),
      fetchCalls(callPage, callPageSize, callStatus, callKeyword),
    ]);
  };

  const handleSend = async (values: NotificationForm) => {
    const payload: SendNotificationPayload = {
      channel: values.channel,
      receiver: values.receiver,
      message: values.message,
    };

    setManualSubmitting(true);
    try {
      await sendNotification(payload);
      messageApi.success("通知发送成功");
      manualForm.resetFields();
      await fetchLogs(1, logPageSize);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "通知发送失败"));
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
      await dispatchScenarioNotification(payload);
      messageApi.success("场景通知发送成功");
      scenarioForm.resetFields();
      await fetchLogs(1, logPageSize);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "场景通知发送失败"));
    } finally {
      setScenarioSubmitting(false);
    }
  };

  const handleUpdateContact = async (values: ContactForm) => {
    setContactSubmitting(true);
    try {
      await updateStudentContact(values.student_id.trim(), {
        student_name: values.student_name?.trim() || undefined,
        guardian_name: values.guardian_name?.trim() || undefined,
        guardian_phone: values.guardian_phone?.trim() || undefined,
        guardian_relation: values.guardian_relation?.trim() || undefined,
      });
      messageApi.success("家长联系人已保存");
      contactForm.resetFields();
      await fetchContacts(1, contactPageSize, contactKeyword);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "保存家长联系人失败"));
    } finally {
      setContactSubmitting(false);
    }
  };

  const handleRetryCall = async (id: string) => {
    setRetryingId(id);
    try {
      await retryOutboundCall(id);
      messageApi.success("外呼已重新发起");
      await fetchCalls(1, callPageSize, callStatus, callKeyword);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "重试外呼失败"));
    } finally {
      setRetryingId(null);
    }
  };

  const notificationColumns: ColumnsType<NotificationLog> = [
    { title: "日志 ID", dataIndex: "id", width: 200 },
    {
      title: "渠道",
      dataIndex: "channel",
      width: 120,
      render: (value: string) => <Tag>{getChannelLabel(value)}</Tag>,
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (value: string) => <Tag color={statusColor(value)}>{getStatusLabel(value)}</Tag>,
    },
    { title: "接收人", dataIndex: "receiver", width: 200, render: (value: string) => value || "-" },
    { title: "消息内容", dataIndex: "message" },
    {
      title: "发送时间",
      dataIndex: "sent_at",
      width: 180,
      render: (value: string) => formatDate(value),
    },
    { title: "错误", dataIndex: "error", width: 220, render: (value?: string) => value || "-" },
  ];

  const contactColumns: ColumnsType<StudentContact> = [
    { title: "学号", dataIndex: "student_id", width: 140 },
    {
      title: "学生姓名",
      dataIndex: "student_name",
      width: 160,
      render: (value: string) => value || "-",
    },
    {
      title: "家长姓名",
      dataIndex: "guardian_name",
      width: 160,
      render: (value: string) => value || "-",
    },
    {
      title: "手机号",
      dataIndex: "guardian_phone",
      width: 180,
      render: (value: string) => value || "-",
    },
    {
      title: "关系",
      dataIndex: "guardian_relation",
      width: 120,
      render: (value: string) => value || "-",
    },
  ];

  const callColumns: ColumnsType<OutboundCall> = [
    { title: "外呼 ID", dataIndex: "id", width: 200 },
    { title: "学号", dataIndex: "student_id", width: 140 },
    { title: "学生", dataIndex: "student_name", width: 120 },
    {
      title: "家长",
      dataIndex: "guardian_name",
      width: 120,
      render: (value: string) => value || "-",
    },
    { title: "手机号", dataIndex: "guardian_phone", width: 160 },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (value: string) => <Tag color={statusColor(value)}>{getStatusLabel(value)}</Tag>,
    },
    {
      title: "供应商",
      dataIndex: "provider",
      width: 100,
      render: (value: string) => <Tag>{getProviderLabel(value)}</Tag>,
    },
    {
      title: "触发方式",
      dataIndex: "trigger_source",
      width: 120,
      render: (value: string) => getTriggerSourceLabel(value),
    },
    {
      title: "请求时间",
      dataIndex: "requested_at",
      width: 180,
      render: (value: string) => formatDate(value),
    },
    { title: "错误原因", dataIndex: "error", width: 220, render: (value?: string) => value || "-" },
    {
      title: "操作",
      key: "actions",
      width: 120,
      render: (_, record) => (
        <Button
          size="small"
          onClick={() => void handleRetryCall(record.id)}
          loading={retryingId === record.id}
        >
          重试外呼
        </Button>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}

      <Space style={{ width: "100%", justifyContent: "space-between" }} align="center">
        <Typography.Title level={3} style={{ marginBottom: 0 }}>
          消息通知与智能外呼
        </Typography.Title>
        <Button icon={<ReloadOutlined />} onClick={() => void refreshAll()}>
          刷新全部
        </Button>
      </Space>

      <Card title="发送通知（微信 / 钉钉）">
        <Form layout="vertical" form={manualForm} onFinish={(values) => void handleSend(values)}>
          <Space align="start" wrap>
            <Form.Item
              label="发送渠道"
              name="channel"
              rules={[{ required: true, message: "请选择发送渠道" }]}
            >
              <Select style={{ width: 160 }} options={channelOptions} />
            </Form.Item>
            <Form.Item
              label="接收人"
              name="receiver"
              rules={[{ required: true, message: "请输入接收人" }]}
            >
              <Input style={{ width: 240 }} placeholder="如：家长通知群" />
            </Form.Item>
            <Form.Item
              label="通知内容"
              name="message"
              rules={[{ required: true, message: "请输入通知内容" }]}
            >
              <Input.TextArea rows={2} style={{ width: 380 }} placeholder="请输入通知内容" />
            </Form.Item>
            <Form.Item label=" " style={{ marginBottom: 0 }}>
              <Button htmlType="submit" type="primary" loading={manualSubmitting}>
                发送通知
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      <Card title="场景通知">
        <Form
          layout="vertical"
          form={scenarioForm}
          onFinish={(values) => void handleScenarioDispatch(values)}
        >
          <Space align="start" wrap>
            <Form.Item
              label="场景"
              name="scenario"
              rules={[{ required: true, message: "请选择场景" }]}
            >
              <Select style={{ width: 220 }} options={scenarioOptions} />
            </Form.Item>
            <Form.Item
              label="发送渠道"
              name="channel"
              rules={[{ required: true, message: "请选择发送渠道" }]}
            >
              <Select style={{ width: 160 }} options={channelOptions} />
            </Form.Item>
            <Form.Item
              label="接收人"
              name="receiver"
              rules={[{ required: true, message: "请输入接收人" }]}
            >
              <Input style={{ width: 220 }} placeholder="如：家长通知群" />
            </Form.Item>
            <Form.Item label="学生姓名" name="student_name">
              <Input style={{ width: 180 }} placeholder="可选" />
            </Form.Item>
            <Form.Item label="去向/地点" name="destination">
              <Select
                allowClear
                showSearch
                style={{ width: 180 }}
                placeholder="可选"
                options={destinationOptions}
                optionFilterProp="label"
              />
            </Form.Item>
            <Form.Item label="复诊时间" name="follow_up_at">
              <Input style={{ width: 220 }} placeholder="可选，如 2026-03-07 14:30" />
            </Form.Item>
            <Form.Item label="备注" name="note">
              <Input.TextArea rows={2} style={{ width: 260 }} placeholder="可选" />
            </Form.Item>
            <Form.Item label=" " style={{ marginBottom: 0 }}>
              <Button htmlType="submit" type="primary" loading={scenarioSubmitting}>
                发送场景通知
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      <Card title="家长联系人维护">
        <Form
          layout="vertical"
          form={contactForm}
          onFinish={(values) => void handleUpdateContact(values)}
        >
          <Space align="start" wrap>
            <Form.Item
              label="学号"
              name="student_id"
              rules={[{ required: true, message: "请输入学号" }]}
            >
              <Input style={{ width: 160 }} placeholder="如：20260001" />
            </Form.Item>
            <Form.Item label="学生姓名" name="student_name">
              <Input style={{ width: 160 }} placeholder="可选" />
            </Form.Item>
            <Form.Item label="家长姓名" name="guardian_name">
              <Input style={{ width: 160 }} placeholder="如：张家长" />
            </Form.Item>
            <Form.Item
              label="家长手机号"
              name="guardian_phone"
              rules={[{ required: true, message: "请输入手机号" }]}
            >
              <Input style={{ width: 180 }} placeholder="如：13800000001" />
            </Form.Item>
            <Form.Item label="关系" name="guardian_relation">
              <Input style={{ width: 120 }} placeholder="如：父亲" />
            </Form.Item>
            <Form.Item label=" " style={{ marginBottom: 0 }}>
              <Button htmlType="submit" type="primary" loading={contactSubmitting}>
                保存联系人
              </Button>
            </Form.Item>
          </Space>
        </Form>

        <Space style={{ marginBottom: 16, marginTop: 8 }}>
          <Input.Search
            allowClear
            placeholder="搜索学号/学生/家长/手机号"
            style={{ width: 280 }}
            onSearch={(value) => {
              setContactPage(1);
              setContactKeyword(value.trim());
            }}
          />
        </Space>

        <Table
          rowKey="student_id"
          columns={contactColumns}
          dataSource={contacts}
          loading={contactsLoading}
          pagination={
            {
              current: contactPage,
              pageSize: contactPageSize,
              total: contactTotal,
              showSizeChanger: true,
              onChange: (page: number, pageSize: number) => {
                setContactPage(page);
                setContactPageSize(pageSize);
              },
            } satisfies TablePaginationConfig
          }
          scroll={{ x: 900 }}
        />
      </Card>

      <Card title="智能外呼追踪">
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            allowClear
            placeholder="按状态筛选"
            style={{ width: 180 }}
            value={callStatus}
            onChange={(value) => {
              setCallPage(1);
              setCallStatus(value);
            }}
            options={[
              { value: "requested", label: "requested" },
              { value: "connected", label: "connected" },
              { value: "failed", label: "failed" },
              { value: "busy", label: "busy" },
              { value: "no_answer", label: "no_answer" },
              { value: "cancelled", label: "cancelled" },
            ]}
          />
          <Input.Search
            allowClear
            placeholder="搜索学号/学生/家长/手机号"
            style={{ width: 280 }}
            onSearch={(value) => {
              setCallPage(1);
              setCallKeyword(value.trim());
            }}
          />
        </Space>

        <Table
          rowKey="id"
          columns={callColumns}
          dataSource={calls}
          loading={callsLoading}
          pagination={
            {
              current: callPage,
              pageSize: callPageSize,
              total: callTotal,
              showSizeChanger: true,
              onChange: (page: number, pageSize: number) => {
                setCallPage(page);
                setCallPageSize(pageSize);
              },
            } satisfies TablePaginationConfig
          }
          scroll={{ x: 1600 }}
        />
      </Card>

      <Card title="通知发送日志">
        <Table
          rowKey="id"
          columns={notificationColumns}
          dataSource={logs}
          loading={logsLoading}
          pagination={
            {
              current: logPage,
              pageSize: logPageSize,
              total: logTotal,
              showSizeChanger: true,
              onChange: (page: number, pageSize: number) => {
                setLogPage(page);
                setLogPageSize(pageSize);
              },
            } satisfies TablePaginationConfig
          }
          scroll={{ x: 1300 }}
        />
      </Card>
    </Space>
  );
}
