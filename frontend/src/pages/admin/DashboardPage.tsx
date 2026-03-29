import {
  AlertOutlined,
  CalendarOutlined,
  EyeOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { Card, Col, Row, Space, Statistic, Typography } from "antd";
import { useEffect, useState } from "react";
import { getOverviewReport, type OverviewReport } from "@/shared/api/reports";
import { getStoredUser } from "@/shared/auth/session";

const defaultReport: OverviewReport = {
  today_visits: 0,
  observation_students: 0,
  stock_warnings: 0,
  due_follow_ups: 0,
};

type StatItem = {
  title: string;
  key: keyof OverviewReport;
  icon: React.ReactNode;
  colorClass: string;
  color: string;
};

const stats: StatItem[] = [
  {
    title: "今日就诊",
    key: "today_visits",
    icon: <TeamOutlined />,
    colorClass: "stat-card--teal",
    color: "#0d9488",
  },
  {
    title: "留观学生",
    key: "observation_students",
    icon: <EyeOutlined />,
    colorClass: "stat-card--sky",
    color: "#0ea5e9",
  },
  {
    title: "库存预警",
    key: "stock_warnings",
    icon: <AlertOutlined />,
    colorClass: "stat-card--amber",
    color: "#f59e0b",
  },
  {
    title: "待复诊",
    key: "due_follow_ups",
    icon: <CalendarOutlined />,
    colorClass: "stat-card--rose",
    color: "#ef4444",
  },
];

export function DashboardPage() {
  const [report, setReport] = useState(defaultReport);
  const [loading, setLoading] = useState(false);
  const currentUser = getStoredUser();

  const today = new Date().toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric",
    weekday: "long",
  });

  useEffect(() => {
    const run = async () => {
      setLoading(true);
      try {
        setReport(await getOverviewReport());
      } finally {
        setLoading(false);
      }
    };
    void run();
  }, []);

  return (
    <Space direction="vertical" size={20} style={{ width: "100%" }}>
      {/* Welcome Banner */}
      <Card className="welcome-banner" bordered={false}>
        <Row align="middle" justify="space-between">
          <Col>
            <Typography.Title level={4} style={{ color: "#fff", marginBottom: 4 }}>
              👋 {currentUser ? `你好，${currentUser.name}` : "欢迎使用"}
            </Typography.Title>
            <Typography.Text style={{ color: "rgba(255,255,255,0.8)", fontSize: 14 }}>
              {today} · 医务室运行概览
            </Typography.Text>
          </Col>
        </Row>
      </Card>

      {/* Stat Cards */}
      <Row gutter={[16, 16]}>
        {stats.map((item) => (
          <Col xs={12} sm={12} md={6} key={item.key}>
            <Card loading={loading} className={`stat-card ${item.colorClass}`}>
              <span className="stat-icon">{item.icon}</span>
              <Statistic
                title={item.title}
                value={report[item.key]}
                valueStyle={{ color: item.color, fontWeight: 700 }}
              />
            </Card>
          </Col>
        ))}
      </Row>
    </Space>
  );
}
