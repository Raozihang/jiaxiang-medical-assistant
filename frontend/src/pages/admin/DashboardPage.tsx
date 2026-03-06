import { Card, Col, Row, Space, Statistic, Typography } from "antd";
import { useEffect, useState } from "react";
import { getOverviewReport, type OverviewReport } from "@/shared/api/reports";

const defaultReport: OverviewReport = {
  today_visits: 0,
  observation_students: 0,
  stock_warnings: 0,
  due_follow_ups: 0,
};

export function DashboardPage() {
  const [report, setReport] = useState(defaultReport);
  const [loading, setLoading] = useState(false);

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
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        管理仪表盘
      </Typography.Title>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="今日就诊" value={report.today_visits} />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="留观学生" value={report.observation_students} />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="库存预警" value={report.stock_warnings} />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="待复诊" value={report.due_follow_ups} />
          </Card>
        </Col>
      </Row>
    </Space>
  );
}
