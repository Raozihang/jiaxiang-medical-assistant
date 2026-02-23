import { Button, Card, InputNumber, message, Progress, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import {
  inboundMedicine,
  listMedicines,
  type Medicine,
  outboundMedicine,
} from "@/shared/api/medicines";

type MedicineRow = {
  id: string;
  name: string;
  stock: number;
  safeStock: number;
  expiryDate: string;
};

function toRow(item: Medicine): MedicineRow {
  return {
    id: item.id,
    name: item.name,
    stock: item.stock,
    safeStock: item.safe_stock,
    expiryDate: new Date(item.expiry_date).toLocaleDateString(),
  };
}

export function MedicinesPage() {
  const [rows, setRows] = useState<MedicineRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [quantity, setQuantity] = useState(10);
  const [messageApi, contextHolder] = message.useMessage();

  const fetchList = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listMedicines({ page: 1, pageSize: 50 });
      setRows(data.items.map(toRow));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchList();
  }, [fetchList]);

  const changeStock = async (id: string, outbound: boolean) => {
    if (outbound) {
      await outboundMedicine({ medicine_id: id, quantity });
      messageApi.success("Outbound completed");
    } else {
      await inboundMedicine({ medicine_id: id, quantity });
      messageApi.success("Inbound completed");
    }
    await fetchList();
  };

  const columns: ColumnsType<MedicineRow> = [
    { title: "Medicine", dataIndex: "name" },
    { title: "Stock", dataIndex: "stock" },
    {
      title: "Stock Status",
      render: (_, row) => {
        const percent = Math.min(100, Math.round((row.stock / row.safeStock) * 100));
        const status = row.stock < row.safeStock ? "exception" : "normal";
        return <Progress percent={percent} status={status} size="small" />;
      },
    },
    { title: "Expiry Date", dataIndex: "expiryDate" },
    {
      title: "Action",
      render: (_, row) => (
        <Space>
          <Button size="small" onClick={() => void changeStock(row.id, false)}>
            Inbound
          </Button>
          <Button size="small" danger onClick={() => void changeStock(row.id, true)}>
            Outbound
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        Medicine Inventory
      </Typography.Title>
      <Card>
        <Space style={{ marginBottom: 12 }}>
          <Typography.Text>Stock change quantity:</Typography.Text>
          <InputNumber
            min={1}
            value={quantity}
            onChange={(value) => setQuantity(Number(value) || 1)}
          />
        </Space>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={rows}
          loading={loading}
          pagination={false}
        />
      </Card>
    </Space>
  );
}
