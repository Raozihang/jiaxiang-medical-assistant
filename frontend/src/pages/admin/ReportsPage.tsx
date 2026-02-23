import { ReloadOutlined } from "@ant-design/icons";
import { Button, Card, Col, Row, Segmented, Space, Statistic, Table, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import {
  getDailyReport,
  getMonthlyReport,
  getWeeklyReport,
  type PeriodReport,
  type ReportRankItem,
  type ReportTrend,
} from "@/shared/api/reports";
import { getErrorMessage } from "@/shared/api/helpers";

const defaultReport: PeriodReport = {
  period: "daily",
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
  raw: null,
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

  useEffect(() => {
    void fetchReport(mode);
  }, [fetchReport, mode]);

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        日/周/月报表
      </Typography.Title>

      <Card
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
            <Statistic title="留观学生" value={report.summary.observationStudents} loading={loading} />
          </Col>
          <Col span={8}>
            <Statistic title="转诊数量" value={report.summary.hospitalReferrals} loading={loading} />
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
    </Space>
  );
}

