import { DeleteOutlined, DownloadOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import {
  Button,
  Card,
  Checkbox,
  Col,
  Form,
  Input,
  Modal,
  message,
  Popconfirm,
  Row,
  Segmented,
  Select,
  Space,
  Statistic,
  Switch,
  Table,
  Tag,
  TimePicker,
  Typography,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import type dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  type ColumnOption,
  createSchedule,
  createTemplate,
  deleteSchedule,
  deleteTemplate,
  downloadScheduleFile,
  exportWithTemplate,
  getColumnOptions,
  listScheduleFiles,
  listSchedules,
  listTemplates,
  type ReportSchedule,
  type ReportTemplate,
  type ScheduledReportFile,
  triggerSchedule,
  updateSchedule,
} from "@/shared/api/report-templates";
import {
  exportReportExcel,
  getDailyReport,
  getMonthlyReport,
  getWeeklyReport,
  type PeriodReport,
  type ReportRankItem,
  type ReportTrend,
} from "@/shared/api/reports";
import { getPeriodLabel } from "@/shared/labels/localization";

const defaultReport: PeriodReport = {
  period: "daily",
  startAt: "",
  endAt: "",
  generatedAt: "",
  summary: {
    totalVisits: 0,
    urgentVisits: 0,
    observationStudents: 0,
    hospitalReferrals: 0,
    returnClassCount: 0,
    stockWarnings: 0,
  },
  trends: [],
  topSymptoms: [],
  topMedicines: [],
  destinationDistribution: {},
  raw: null,
};

function formatDate(value: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

const trendColumns: ColumnsType<ReportTrend> = [
  { title: "统计维度", dataIndex: "label" },
  { title: "就诊量", dataIndex: "visits" },
  { title: "紧急数", dataIndex: "urgent" },
  { title: "留观数", dataIndex: "observation" },
];

const rankColumns: ColumnsType<ReportRankItem> = [
  { title: "名称", dataIndex: "name" },
  { title: "次数", dataIndex: "count", width: 120 },
];

type ReportMode = "daily" | "weekly" | "monthly";

export function ReportsPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [mode, setMode] = useState<ReportMode>("daily");
  const [report, setReport] = useState<PeriodReport>(defaultReport);
  const [loading, setLoading] = useState(false);
  const [exporting, setExporting] = useState(false);

  // Template state
  const [templates, setTemplates] = useState<ReportTemplate[]>([]);
  const [columnOptions, setColumnOptions] = useState<ColumnOption[]>([]);
  const [showTplModal, setShowTplModal] = useState(false);
  const [tplForm] = Form.useForm();

  // Schedule state
  const [schedules, setSchedules] = useState<ReportSchedule[]>([]);
  const [showSchedModal, setShowSchedModal] = useState(false);
  const [schedForm] = Form.useForm();
  const [runningScheduleId, setRunningScheduleId] = useState<string | null>(null);
  const [historyScheduleId, setHistoryScheduleId] = useState<string | null>(null);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyFiles, setHistoryFiles] = useState<ScheduledReportFile[]>([]);

  const handleExport = useCallback(
    async (targetMode: ReportMode) => {
      setExporting(true);
      try {
        await exportReportExcel(targetMode);
        messageApi.success("Excel 下载成功");
      } catch (error) {
        messageApi.error(getErrorMessage(error, "导出失败"));
      } finally {
        setExporting(false);
      }
    },
    [messageApi],
  );

  const fetchReport = useCallback(
    async (targetMode: ReportMode) => {
      setLoading(true);
      try {
        const nextReport =
          targetMode === "daily"
            ? await getDailyReport()
            : targetMode === "weekly"
              ? await getWeeklyReport()
              : await getMonthlyReport();
        setReport(nextReport);
      } catch (error) {
        messageApi.error(getErrorMessage(error, "报表获取失败"));
      } finally {
        setLoading(false);
      }
    },
    [messageApi],
  );

  const fetchTemplates = useCallback(async () => {
    try {
      const [tpls, cols] = await Promise.all([listTemplates(), getColumnOptions()]);
      setTemplates(tpls);
      setColumnOptions(cols);
    } catch {
      /* ignore */
    }
  }, []);

  const fetchSchedules = useCallback(async () => {
    try {
      setSchedules(await listSchedules());
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    void fetchReport(mode);
    void fetchTemplates();
    void fetchSchedules();
  }, [fetchReport, fetchTemplates, fetchSchedules, mode]);

  // ---- Template actions ----

  const handleCreateTemplate = async (values: {
    name: string;
    period: string;
    columns: string[];
    title?: string;
  }) => {
    try {
      await createTemplate(values);
      messageApi.success("模板创建成功");
      setShowTplModal(false);
      tplForm.resetFields();
      void fetchTemplates();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "创建失败"));
    }
  };

  const handleDeleteTemplate = async (id: string) => {
    try {
      await deleteTemplate(id);
      messageApi.success("模板已删除");
      void fetchTemplates();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "删除失败"));
    }
  };

  const handleExportWithTemplate = async (id: string) => {
    try {
      await exportWithTemplate(id);
      messageApi.success("导出成功");
    } catch (error) {
      messageApi.error(getErrorMessage(error, "导出失败"));
    }
  };

  // ---- Schedule actions ----

  const handleCreateSchedule = async (values: { template_id: string; time: dayjs.Dayjs }) => {
    try {
      const cronExpr = values.time.format("HH:mm");
      await createSchedule({ template_id: values.template_id, cron_expr: cronExpr });
      messageApi.success("定时任务创建成功");
      setShowSchedModal(false);
      schedForm.resetFields();
      void fetchSchedules();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "创建失败"));
    }
  };

  const handleToggleSchedule = async (id: string, enabled: boolean) => {
    try {
      await updateSchedule(id, { enabled });
      void fetchSchedules();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "更新失败"));
    }
  };

  const handleDeleteSchedule = async (id: string) => {
    try {
      await deleteSchedule(id);
      messageApi.success("定时任务已删除");
      void fetchSchedules();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "删除失败"));
    }
  };

  const handleRunSchedule = async (id: string) => {
    setRunningScheduleId(id);
    try {
      await triggerSchedule(id);
      messageApi.success("已立即执行并下载 Excel");
      void fetchSchedules();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "执行失败"));
    } finally {
      setRunningScheduleId(null);
    }
  };

  const openHistoryModal = async (id: string) => {
    setHistoryScheduleId(id);
    setHistoryLoading(true);
    try {
      setHistoryFiles(await listScheduleFiles(id));
    } catch (error) {
      messageApi.error(getErrorMessage(error, "获取历史文件失败"));
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleDownloadHistoryFile = async (scheduleId: string, fileName: string) => {
    try {
      await downloadScheduleFile(scheduleId, fileName);
      messageApi.success("历史文件下载成功");
    } catch (error) {
      messageApi.error(getErrorMessage(error, "下载失败"));
    }
  };

  // ---- Template table columns ----

  const templateColumns: ColumnsType<ReportTemplate> = [
    { title: "模板名称", dataIndex: "name", width: 160 },
    {
      title: "报表类型",
      dataIndex: "period",
      width: 100,
      render: (p: string) => <Tag color="blue">{getPeriodLabel(p)}</Tag>,
    },
    {
      title: "包含列",
      dataIndex: "columns",
      render: (cols: string[]) => {
        const colMap = Object.fromEntries(columnOptions.map((c) => [c.key, c.label]));
        return cols.map((c) => colMap[c] ?? "未知列").join("、");
      },
    },
    { title: "创建时间", dataIndex: "created_at", width: 180, render: formatDate },
    {
      title: "操作",
      width: 160,
      render: (_: unknown, record: ReportTemplate) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => void handleExportWithTemplate(record.id)}
          >
            导出
          </Button>
          <Popconfirm title="确认删除?" onConfirm={() => void handleDeleteTemplate(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // ---- Schedule table columns ----

  const scheduleColumns: ColumnsType<ReportSchedule> = [
    {
      title: "关联模板",
      dataIndex: "template_id",
      render: (id: string) => {
        const tpl = templates.find((t) => t.id === id);
        return tpl?.name ?? id.slice(0, 8);
      },
    },
    { title: "执行时间", dataIndex: "cron_expr", width: 120 },
    {
      title: "状态",
      dataIndex: "enabled",
      width: 100,
      render: (enabled: boolean, record: ReportSchedule) => (
        <Switch
          checked={enabled}
          onChange={(checked) => void handleToggleSchedule(record.id, checked)}
          checkedChildren="启用"
          unCheckedChildren="停用"
        />
      ),
    },
    {
      title: "下次执行",
      dataIndex: "next_run_at",
      width: 180,
      render: (v: string | null) => (v ? formatDate(v) : "-"),
    },
    {
      title: "上次执行",
      dataIndex: "last_run_at",
      width: 180,
      render: (v: string | null) => (v ? formatDate(v) : "-"),
    },
    {
      title: "操作",
      width: 220,
      render: (_: unknown, record: ReportSchedule) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<DownloadOutlined />}
            loading={runningScheduleId === record.id}
            onClick={() => void handleRunSchedule(record.id)}
          >
            立即执行
          </Button>
          <Button type="link" size="small" onClick={() => void openHistoryModal(record.id)}>
            历史文件
          </Button>
          <Popconfirm title="确认删除?" onConfirm={() => void handleDeleteSchedule(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const historyColumns: ColumnsType<ScheduledReportFile> = [
    { title: "文件名", dataIndex: "name" },
    {
      title: "大小",
      dataIndex: "size_bytes",
      width: 120,
      render: (value: number) => `${(value / 1024).toFixed(value >= 1024 ? 1 : 0)} KB`,
    },
    { title: "修改时间", dataIndex: "modified_at", width: 180, render: formatDate },
    {
      title: "操作",
      width: 100,
      render: (_: unknown, record: ScheduledReportFile) =>
        historyScheduleId ? (
          <Button
            type="link"
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => void handleDownloadHistoryFile(historyScheduleId, record.name)}
          >
            下载
          </Button>
        ) : null,
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        报表中心
      </Typography.Title>

      {/* ---- Quick report ---- */}
      <Card
        title="快速报表"
        extra={
          <Space>
            <Segmented<ReportMode>
              value={mode}
              options={[
                { value: "daily", label: "日报" },
                { value: "weekly", label: "周报" },
                { value: "monthly", label: "月报" },
              ]}
              onChange={(value) => setMode(value)}
            />
            <Button icon={<ReloadOutlined />} onClick={() => void fetchReport(mode)}>
              刷新
            </Button>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              loading={exporting}
              onClick={() => void handleExport(mode)}
            >
              导出 Excel
            </Button>
          </Space>
        }
      >
        <Typography.Text type="secondary">
          报表周期：{report.period} | 更新时间：{formatDate(report.generatedAt)}
        </Typography.Text>

        <Row gutter={[16, 16]} style={{ marginTop: 12 }}>
          <Col span={8}>
            <Statistic title="就诊总量" value={report.summary.totalVisits} loading={loading} />
          </Col>
          <Col span={8}>
            <Statistic title="紧急就诊" value={report.summary.urgentVisits} loading={loading} />
          </Col>
          <Col span={8}>
            <Statistic
              title="留观学生"
              value={report.summary.observationStudents}
              loading={loading}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="转诊数量"
              value={report.summary.hospitalReferrals}
              loading={loading}
            />
          </Col>
          <Col span={8}>
            <Statistic title="返班数量" value={report.summary.returnClassCount} loading={loading} />
          </Col>
          <Col span={8}>
            <Statistic title="库存预警" value={report.summary.stockWarnings} loading={loading} />
          </Col>
        </Row>
      </Card>

      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card title="趋势统计" loading={loading}>
            <Table
              rowKey="label"
              columns={trendColumns}
              dataSource={report.trends}
              pagination={{ pageSize: 6 }}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="高频症状" loading={loading}>
            <Table
              rowKey="name"
              columns={rankColumns}
              dataSource={report.topSymptoms}
              pagination={{ pageSize: 6 }}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="高频用药" loading={loading}>
            <Table
              rowKey="name"
              columns={rankColumns}
              dataSource={report.topMedicines}
              pagination={{ pageSize: 6 }}
            />
          </Card>
        </Col>
      </Row>

      {/* ---- Templates ---- */}
      <Card
        title="报表模板"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setShowTplModal(true)}>
            新建模板
          </Button>
        }
      >
        <Table
          rowKey="id"
          columns={templateColumns}
          dataSource={templates}
          pagination={false}
          locale={{ emptyText: "暂无模板，点击「新建模板」创建" }}
        />
      </Card>

      {/* ---- Schedules ---- */}
      <Card
        title="定时任务"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            disabled={templates.length === 0}
            onClick={() => setShowSchedModal(true)}
          >
            新建定时任务
          </Button>
        }
      >
        <Table
          rowKey="id"
          columns={scheduleColumns}
          dataSource={schedules}
          pagination={false}
          locale={{ emptyText: "暂无定时任务" }}
        />
      </Card>

      {/* ---- Create Template Modal ---- */}
      <Modal
        title="新建报表模板"
        open={showTplModal}
        onCancel={() => setShowTplModal(false)}
        onOk={() => tplForm.submit()}
        okText="创建"
        cancelText="取消"
      >
        <Form
          form={tplForm}
          layout="vertical"
          onFinish={handleCreateTemplate}
          initialValues={{ period: "daily", columns: columnOptions.map((c) => c.key) }}
        >
          <Form.Item
            name="name"
            label="模板名称"
            rules={[{ required: true, message: "请输入模板名称" }]}
          >
            <Input placeholder="例如：每日简报" />
          </Form.Item>
          <Form.Item name="period" label="报表类型" rules={[{ required: true }]}>
            <Select
              options={[
                { value: "daily", label: "日报" },
                { value: "weekly", label: "周报" },
                { value: "monthly", label: "月报" },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="columns"
            label="包含列"
            rules={[{ required: true, message: "请至少选择一列" }]}
          >
            <Checkbox.Group
              options={columnOptions.map((c) => ({ label: c.label, value: c.key }))}
            />
          </Form.Item>
          <Form.Item name="title" label="自定义标题">
            <Input placeholder="留空则使用默认标题" />
          </Form.Item>
        </Form>
      </Modal>

      {/* ---- Create Schedule Modal ---- */}
      <Modal
        title="新建定时任务"
        open={showSchedModal}
        onCancel={() => setShowSchedModal(false)}
        onOk={() => schedForm.submit()}
        okText="创建"
        cancelText="取消"
      >
        <Form form={schedForm} layout="vertical" onFinish={handleCreateSchedule}>
          <Form.Item
            name="template_id"
            label="关联模板"
            rules={[{ required: true, message: "请选择模板" }]}
          >
            <Select
              placeholder="选择报表模板"
              options={templates.map((t) => ({
                value: t.id,
                label: `${t.name} (${getPeriodLabel(t.period)})`,
              }))}
            />
          </Form.Item>
          <Form.Item
            name="time"
            label="执行时间"
            rules={[{ required: true, message: "请选择时间" }]}
          >
            <TimePicker format="HH:mm" style={{ width: "100%" }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="历史导出文件"
        open={historyScheduleId !== null}
        onCancel={() => {
          setHistoryScheduleId(null);
          setHistoryFiles([]);
        }}
        footer={null}
        width={820}
      >
        <Table
          rowKey="name"
          loading={historyLoading}
          columns={historyColumns}
          dataSource={historyFiles}
          pagination={false}
          locale={{ emptyText: "暂无历史导出文件" }}
        />
      </Modal>
    </Space>
  );
}
