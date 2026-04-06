import { Button, Card, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { listVisits, type Visit } from "@/shared/api/visits";

type VisitRow = {
  id: string;
  studentName: string;
  className: string;
  symptom: string;
  level: "normal" | "urgent";
  createdAt: string;
};

function toRow(visit: Visit): VisitRow {
  const symptoms = visit.symptoms.join(", ");
  return {
    id: visit.id,
    studentName: visit.student_name,
    className: visit.class_name,
    symptom: symptoms || "-",
    level: visit.destination === "urgent" ? "urgent" : "normal",
    createdAt: new Date(visit.created_at).toLocaleString(),
  };
}

export function VisitsPage() {
  const navigate = useNavigate();
  const [rows, setRows] = useState<VisitRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  const fetchList = useCallback(
    async (targetPage: number) => {
      setLoading(true);
      try {
        const data = await listVisits({ page: targetPage, pageSize });
        setRows(data.items.map(toRow));
        setPage(data.page);
        setTotal(data.total);
      } finally {
        setLoading(false);
      }
    },
    [pageSize],
  );

  useEffect(() => {
    void fetchList(1);
  }, [fetchList]);

  const columns: ColumnsType<VisitRow> = [
    { title: "学生姓名", dataIndex: "studentName" },
    { title: "班级", dataIndex: "className" },
    { title: "症状", dataIndex: "symptom" },
    {
      title: "优先级",
      dataIndex: "level",
      render: (value: VisitRow["level"]) =>
        value === "urgent" ? <Tag color="red">紧急</Tag> : <Tag color="blue">普通</Tag>,
    },
    { title: "创建时间", dataIndex: "createdAt" },
    {
      title: "操作",
      render: (_, row) => (
        <Button type="link" onClick={() => navigate(`/doctor/visit/${row.id}`)}>
          查看
        </Button>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        就诊队列
      </Typography.Title>
      <Card>
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
              onChange: (targetPage: number) => {
                void fetchList(targetPage);
              },
            } as TablePaginationConfig
          }
        />
      </Card>
    </Space>
  );
}
