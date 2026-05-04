import {
  AlertOutlined,
  BarChartOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  MedicineBoxOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Button,
  Card,
  Col,
  message,
  Progress,
  Row,
  Skeleton,
  Space,
  Statistic,
  Tag,
  Typography,
} from "antd";
import { useCallback, useEffect, useMemo, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  type DestinationDistribution,
  getMonthlyReport,
  getOverviewReport,
  getWeeklyReport,
  type OverviewReport,
  type PeriodReport,
} from "@/shared/api/reports";
import { getDestinationLabel } from "@/shared/labels/localization";

const defaultReport: OverviewReport = {
  today_visits: 0,
  observation_students: 0,
  stock_warnings: 0,
  due_follow_ups: 0,
};

const emptyPeriodReport: PeriodReport = {
  period: "",
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

const destinationColors: Record<string, string> = {
  observation: "#f59e0b",
  return_class: "#0d9488",
  back_to_class: "#0d9488",
  classroom: "#0d9488",
  hospital: "#ef4444",
  urgent: "#dc2626",
  referred: "#f97316",
  leave_school: "#8b5cf6",
  back_to_dorm: "#0ea5e9",
  home: "#8b5cf6",
  unknown: "#94a3b8",
};

function formatDateTime(value: string) {
  if (!value) {
    return "暂无更新";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function percent(value: number, total: number) {
  if (total <= 0) {
    return 0;
  }
  return Math.round((value / total) * 100);
}

function destinationRows(distribution: DestinationDistribution) {
  return Object.entries(distribution)
    .map(([key, count]) => ({
      key,
      count,
      label: getDestinationLabel(key),
      color: destinationColors[key] ?? "#64748b",
    }))
    .sort((a, b) => b.count - a.count);
}

function sumDistribution(distribution: DestinationDistribution) {
  return Object.values(distribution).reduce((sum, item) => sum + item, 0);
}

function MiniTrendChart({ report }: { report: PeriodReport }) {
  const items =
    report.trends.length > 0
      ? report.trends
      : [
          {
            label: "本期",
            visits: report.summary.totalVisits,
            urgent: report.summary.urgentVisits,
            observation: report.summary.observationStudents,
          },
        ];
  const maxValue = Math.max(...items.map((item) => item.visits), 1);

  return (
    <div className="dashboard-trend-chart" role="img" aria-label="就诊趋势图">
      {items.slice(-8).map((item) => {
        const height = Math.max(10, Math.round((item.visits / maxValue) * 100));
        return (
          <div className="dashboard-trend-item" key={item.label}>
            <div className="dashboard-trend-bars">
              <span
                className="dashboard-trend-bar dashboard-trend-bar--observation"
                style={{
                  height: `${Math.max(8, Math.round((item.observation / maxValue) * 100))}%`,
                }}
              />
              <span
                className="dashboard-trend-bar dashboard-trend-bar--visits"
                style={{ height: `${height}%` }}
              />
              <span
                className="dashboard-trend-bar dashboard-trend-bar--urgent"
                style={{ height: `${Math.max(8, Math.round((item.urgent / maxValue) * 100))}%` }}
              />
            </div>
            <Typography.Text ellipsis className="dashboard-trend-label">
              {item.label}
            </Typography.Text>
          </div>
        );
      })}
    </div>
  );
}

function DestinationPanel({ report }: { report: PeriodReport }) {
  const rows = destinationRows(report.destinationDistribution);
  const total = sumDistribution(report.destinationDistribution) || report.summary.totalVisits;

  if (rows.length === 0) {
    return (
      <div className="dashboard-empty-panel">
        <DatabaseOutlined />
        <Typography.Text type="secondary">暂无去向分布数据</Typography.Text>
      </div>
    );
  }

  return (
    <Space direction="vertical" size={14} style={{ width: "100%" }}>
      {rows.map((row) => (
        <div className="dashboard-distribution-row" key={row.key}>
          <Space style={{ width: "100%", justifyContent: "space-between" }}>
            <Space size={8}>
              <span className="dashboard-color-dot" style={{ backgroundColor: row.color }} />
              <Typography.Text>{row.label}</Typography.Text>
            </Space>
            <Typography.Text strong>{row.count}</Typography.Text>
          </Space>
          <Progress
            percent={percent(row.count, total)}
            showInfo={false}
            strokeColor={row.color}
            trailColor="#eef2f7"
          />
        </div>
      ))}
    </Space>
  );
}

function RankingList({
  title,
  items,
}: {
  title: string;
  items: Array<{ name: string; count: number }>;
}) {
  return (
    <Card title={title} className="dashboard-card">
      {items.length === 0 ? (
        <div className="dashboard-empty-panel dashboard-empty-panel--compact">
          <Typography.Text type="secondary">暂无排行数据</Typography.Text>
        </div>
      ) : (
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          {items.slice(0, 5).map((item, index) => (
            <div className="dashboard-rank-row" key={item.name}>
              <Tag color={index === 0 ? "red" : index === 1 ? "orange" : "blue"}>{index + 1}</Tag>
              <Typography.Text ellipsis style={{ flex: 1 }}>
                {item.name}
              </Typography.Text>
              <Typography.Text strong>{item.count}</Typography.Text>
            </div>
          ))}
        </Space>
      )}
    </Card>
  );
}

export function DashboardPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [report, setReport] = useState(defaultReport);
  const [weeklyReport, setWeeklyReport] = useState(emptyPeriodReport);
  const [monthlyReport, setMonthlyReport] = useState(emptyPeriodReport);
  const [loading, setLoading] = useState(false);

  const fetchDashboard = useCallback(async () => {
    setLoading(true);
    try {
      const [overview, weekly, monthly] = await Promise.all([
        getOverviewReport(),
        getWeeklyReport(),
        getMonthlyReport(),
      ]);
      setReport(overview);
      setWeeklyReport(weekly);
      setMonthlyReport(monthly);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "管理仪表盘数据获取失败"));
    } finally {
      setLoading(false);
    }
  }, [messageApi]);

  useEffect(() => {
    void fetchDashboard();
  }, [fetchDashboard]);

  const riskTotal = report.observation_students + report.stock_warnings + report.due_follow_ups;
  const observationRate = percent(
    weeklyReport.summary.observationStudents,
    weeklyReport.summary.totalVisits,
  );
  const referralRate = percent(
    weeklyReport.summary.hospitalReferrals,
    weeklyReport.summary.totalVisits,
  );
  const returnRate = percent(
    weeklyReport.summary.returnClassCount,
    weeklyReport.summary.totalVisits,
  );

  const taskItems = useMemo(
    () => [
      {
        label: "留观学生",
        value: report.observation_students,
        tone: "orange",
        text: report.observation_students > 0 ? "需要持续观察体温与症状变化" : "当前无留观压力",
      },
      {
        label: "待复诊",
        value: report.due_follow_ups,
        tone: "blue",
        text: report.due_follow_ups > 0 ? "建议校医今日完成回访闭环" : "复诊任务已清空",
      },
      {
        label: "库存预警",
        value: report.stock_warnings,
        tone: "red",
        text: report.stock_warnings > 0 ? "请后勤优先核对低库存药品" : "常用药库存平稳",
      },
    ],
    [report],
  );

  return (
    <Space direction="vertical" size={18} style={{ width: "100%" }}>
      {contextHolder}
      <div className="dashboard-page-heading">
        <div>
          <Typography.Title level={3} style={{ marginBottom: 4 }}>
            管理仪表盘
          </Typography.Title>
          <Typography.Text type="secondary">
            汇总今日医务室运行、学生流向、库存风险与周期趋势
          </Typography.Text>
        </div>
        <Space>
          <Tag color={riskTotal > 0 ? "orange" : "green"}>
            {riskTotal > 0 ? `${riskTotal} 项待关注` : "运行平稳"}
          </Tag>
          <Button icon={<ReloadOutlined />} loading={loading} onClick={() => void fetchDashboard()}>
            刷新
          </Button>
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} xl={6}>
          <Card loading={loading} className="stat-card stat-card--teal">
            <MedicineBoxOutlined className="stat-icon" />
            <Statistic title="今日就诊" value={report.today_visits} suffix="人次" />
          </Card>
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <Card loading={loading} className="stat-card stat-card--amber">
            <TeamOutlined className="stat-icon" />
            <Statistic title="留观学生" value={report.observation_students} suffix="人" />
          </Card>
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <Card loading={loading} className="stat-card stat-card--rose">
            <AlertOutlined className="stat-icon" />
            <Statistic title="库存预警" value={report.stock_warnings} suffix="项" />
          </Card>
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <Card loading={loading} className="stat-card stat-card--sky">
            <ClockCircleOutlined className="stat-icon" />
            <Statistic title="待复诊" value={report.due_follow_ups} suffix="人" />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={15}>
          <Card
            className="dashboard-card"
            title="本周就诊趋势"
            extra={
              <Typography.Text type="secondary">
                更新于 {formatDateTime(weeklyReport.generatedAt)}
              </Typography.Text>
            }
          >
            {loading ? (
              <Skeleton active paragraph={{ rows: 6 }} />
            ) : (
              <>
                <MiniTrendChart report={weeklyReport} />
                <Row gutter={[12, 12]} className="dashboard-rate-grid">
                  <Col xs={24} md={8}>
                    <Statistic
                      title="本周就诊总量"
                      value={weeklyReport.summary.totalVisits}
                      suffix="人次"
                    />
                  </Col>
                  <Col xs={24} md={8}>
                    <Statistic title="留观占比" value={observationRate} suffix="%" />
                  </Col>
                  <Col xs={24} md={8}>
                    <Statistic title="转诊占比" value={referralRate} suffix="%" />
                  </Col>
                </Row>
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} xl={9}>
          <Card
            className="dashboard-card"
            title="学生去向分布"
            extra={<SafetyCertificateOutlined />}
          >
            {loading ? (
              <Skeleton active paragraph={{ rows: 5 }} />
            ) : (
              <DestinationPanel report={weeklyReport} />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card className="dashboard-card" title="今日待办">
            <Space direction="vertical" size={14} style={{ width: "100%" }}>
              {taskItems.map((item) => (
                <div className="dashboard-task-row" key={item.label}>
                  <Tag color={item.tone}>{item.label}</Tag>
                  <div>
                    <Typography.Text strong>{item.value}</Typography.Text>
                    <Typography.Text type="secondary"> {item.text}</Typography.Text>
                  </div>
                </div>
              ))}
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card className="dashboard-card" title="本月运行摘要" extra={<BarChartOutlined />}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>
              <Statistic
                title="月累计就诊"
                value={monthlyReport.summary.totalVisits}
                suffix="人次"
                loading={loading}
              />
              <div>
                <Space style={{ width: "100%", justifyContent: "space-between" }}>
                  <Typography.Text type="secondary">返回班级占比</Typography.Text>
                  <Typography.Text strong>
                    {percent(
                      monthlyReport.summary.returnClassCount,
                      monthlyReport.summary.totalVisits,
                    )}
                    %
                  </Typography.Text>
                </Space>
                <Progress
                  percent={percent(
                    monthlyReport.summary.returnClassCount,
                    monthlyReport.summary.totalVisits,
                  )}
                  showInfo={false}
                  strokeColor="#0d9488"
                />
              </div>
              <div>
                <Space style={{ width: "100%", justifyContent: "space-between" }}>
                  <Typography.Text type="secondary">本周返回班级占比</Typography.Text>
                  <Typography.Text strong>{returnRate}%</Typography.Text>
                </Space>
                <Progress percent={returnRate} showInfo={false} strokeColor="#0ea5e9" />
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <RankingList title="本月高频症状" items={monthlyReport.topSymptoms} />
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <RankingList title="本月高频用药" items={monthlyReport.topMedicines} />
        </Col>
        <Col xs={24} lg={12}>
          <Card className="dashboard-card" title="本周管理重点">
            <Space direction="vertical" size={10}>
              <Typography.Text>优先关注留观、复诊、库存三类需要闭环的事项。</Typography.Text>
              <Typography.Text type="secondary">
                高频症状和用药变化可用于晨检提醒、班级健康沟通和常用药补货判断。
              </Typography.Text>
            </Space>
          </Card>
        </Col>
      </Row>
    </Space>
  );
}
