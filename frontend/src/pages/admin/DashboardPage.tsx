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
        Management Dashboard
      </Typography.Title>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="Today Visits" value={report.today_visits} />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="Observation Students" value={report.observation_students} />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="Stock Warnings" value={report.stock_warnings} />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic title="Due Follow-ups" value={report.due_follow_ups} />
          </Card>
        </Col>
      </Row>
    </Space>
  );
}
