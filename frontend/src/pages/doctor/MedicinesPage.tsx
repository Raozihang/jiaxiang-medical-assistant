import {
  Button,
  Card,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Modal,
  Progress,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import dayjs, { type Dayjs } from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  createMedicine,
  inboundMedicine,
  listMedicines,
  type Medicine,
  outboundMedicine,
  updateMedicineInventory,
} from "@/shared/api/medicines";

type MedicineRow = {
  id: string;
  name: string;
  specification: string;
  stock: number;
  safeStock: number;
  expiryDate: string;
  warnings: string[];
  isLowStock: boolean;
  isExpiringSoon: boolean;
};

type CreateMedicineForm = {
  name: string;
  specification: string;
  stock: number;
  safeStock: number;
  expiryDate: Dayjs;
};

type EditInventoryForm = {
  stock: number;
  safeStock: number;
};

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString();
}

function toRow(item: Medicine): MedicineRow {
  return {
    id: item.id,
    name: item.name,
    specification: item.specification,
    stock: item.stock,
    safeStock: item.safe_stock,
    expiryDate: item.expiry_date,
    warnings: item.warnings ?? [],
    isLowStock: item.is_low_stock,
    isExpiringSoon: item.is_expiring_soon,
  };
}

export function MedicinesPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [createForm] = Form.useForm<CreateMedicineForm>();
  const [editForm] = Form.useForm<EditInventoryForm>();

  const [rows, setRows] = useState<MedicineRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [actionLoadingKey, setActionLoadingKey] = useState<string | null>(null);
  const [editingRow, setEditingRow] = useState<MedicineRow | null>(null);
  const [quantityById, setQuantityById] = useState<Record<string, number>>({});

  const fetchList = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listMedicines({ page: 1, pageSize: 100 });
      setRows(data.items.map(toRow));
    } catch (error) {
      messageApi.error(getErrorMessage(error, "获取药品库存失败"));
    } finally {
      setLoading(false);
    }
  }, [messageApi]);

  useEffect(() => {
    void fetchList();
  }, [fetchList]);

  const getQuantity = (id: string) => quantityById[id] ?? 1;

  const handleCreateMedicine = async (values: CreateMedicineForm) => {
    setSubmitting(true);
    try {
      await createMedicine({
        name: values.name.trim(),
        specification: values.specification.trim(),
        stock: values.stock,
        safe_stock: values.safeStock,
        expiry_date: values.expiryDate.format("YYYY-MM-DD"),
      });
      messageApi.success("药品已添加");
      createForm.resetFields();
      createForm.setFieldsValue({
        stock: 0,
        safeStock: 10,
        expiryDate: dayjs().add(1, "year"),
      });
      await fetchList();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "添加药品失败"));
    } finally {
      setSubmitting(false);
    }
  };

  const handleChangeStock = async (row: MedicineRow, outbound: boolean) => {
    const quantity = getQuantity(row.id);
    setActionLoadingKey(`${row.id}-${outbound ? "out" : "in"}`);
    try {
      if (outbound) {
        await outboundMedicine({ medicine_id: row.id, quantity });
        messageApi.success("出库完成");
      } else {
        await inboundMedicine({ medicine_id: row.id, quantity });
        messageApi.success("入库完成");
      }
      await fetchList();
    } catch (error) {
      messageApi.error(getErrorMessage(error, outbound ? "出库失败" : "入库失败"));
    } finally {
      setActionLoadingKey(null);
    }
  };

  const openEditModal = (row: MedicineRow) => {
    setEditingRow(row);
    editForm.setFieldsValue({
      stock: row.stock,
      safeStock: row.safeStock,
    });
  };

  const handleSaveInventory = async (values: EditInventoryForm) => {
    if (!editingRow) {
      return;
    }

    setSubmitting(true);
    try {
      await updateMedicineInventory(editingRow.id, {
        stock: values.stock,
        safe_stock: values.safeStock,
      });
      messageApi.success("库存已更新");
      setEditingRow(null);
      editForm.resetFields();
      await fetchList();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "库存更新失败"));
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ColumnsType<MedicineRow> = [
    {
      title: "药品名称",
      dataIndex: "name",
      render: (_, row) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>{row.name}</Typography.Text>
          <Typography.Text type="secondary">{row.specification}</Typography.Text>
        </Space>
      ),
    },
    {
      title: "当前库存",
      dataIndex: "stock",
      width: 110,
    },
    {
      title: "安全库存",
      dataIndex: "safeStock",
      width: 110,
    },
    {
      title: "库存状态",
      key: "status",
      width: 220,
      render: (_, row) => {
        const percent =
          row.safeStock > 0 ? Math.min(100, Math.round((row.stock / row.safeStock) * 100)) : 100;
        return (
          <Space direction="vertical" size={4} style={{ width: "100%" }}>
            <Progress percent={percent} status={row.isLowStock ? "exception" : "success"} size="small" />
            <Space size={4} wrap>
              {row.isLowStock ? <Tag color="red">低库存</Tag> : <Tag color="green">库存正常</Tag>}
              {row.isExpiringSoon ? <Tag color="orange">临近效期</Tag> : null}
            </Space>
          </Space>
        );
      },
    },
    {
      title: "有效期",
      dataIndex: "expiryDate",
      width: 140,
      render: (value: string, row) => (
        <Typography.Text type={row.isExpiringSoon ? "danger" : undefined}>{formatDate(value)}</Typography.Text>
      ),
    },
    {
      title: "预警说明",
      dataIndex: "warnings",
      render: (warnings: string[]) =>
        warnings.length > 0 ? warnings.join("；") : <Typography.Text type="secondary">无</Typography.Text>,
    },
    {
      title: "数量管理",
      key: "actions",
      width: 280,
      render: (_, row) => (
        <Space wrap>
          <InputNumber
            min={1}
            value={getQuantity(row.id)}
            onChange={(value) =>
              setQuantityById((current) => ({
                ...current,
                [row.id]: Number(value) > 0 ? Number(value) : 1,
              }))
            }
          />
          <Button
            size="small"
            onClick={() => void handleChangeStock(row, false)}
            loading={actionLoadingKey === `${row.id}-in`}
          >
            入库
          </Button>
          <Button
            size="small"
            danger
            onClick={() => void handleChangeStock(row, true)}
            loading={actionLoadingKey === `${row.id}-out`}
          >
            出库
          </Button>
          <Button size="small" type="link" onClick={() => openEditModal(row)}>
            直接改库存
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        药品库存管理
      </Typography.Title>

      <Card title="添加药品">
        <Form
          form={createForm}
          layout="vertical"
          onFinish={(values) => void handleCreateMedicine(values)}
          initialValues={{
            stock: 0,
            safeStock: 10,
            expiryDate: dayjs().add(1, "year"),
          }}
        >
          <Space align="start" wrap>
            <Form.Item
              label="药品名称"
              name="name"
              rules={[{ required: true, message: "请输入药品名称" }]}
            >
              <Input style={{ width: 220 }} placeholder="如：布洛芬片" />
            </Form.Item>
            <Form.Item
              label="规格"
              name="specification"
              rules={[{ required: true, message: "请输入规格" }]}
            >
              <Input style={{ width: 180 }} placeholder="如：0.2g*24片" />
            </Form.Item>
            <Form.Item
              label="初始库存"
              name="stock"
              rules={[{ required: true, message: "请输入初始库存" }]}
            >
              <InputNumber min={0} style={{ width: 120 }} />
            </Form.Item>
            <Form.Item
              label="安全库存"
              name="safeStock"
              rules={[{ required: true, message: "请输入安全库存" }]}
            >
              <InputNumber min={0} style={{ width: 120 }} />
            </Form.Item>
            <Form.Item
              label="有效期"
              name="expiryDate"
              rules={[{ required: true, message: "请选择有效期" }]}
            >
              <DatePicker style={{ width: 160 }} />
            </Form.Item>
            <Form.Item label=" " style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={submitting}>
                添加药品
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      <Card title="库存列表">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={rows}
          loading={loading}
          pagination={false}
          locale={{ emptyText: "暂无药品，请先添加药品" }}
          scroll={{ x: 1100 }}
        />
      </Card>

      <Modal
        title={editingRow ? `编辑库存：${editingRow.name}` : "编辑库存"}
        open={editingRow !== null}
        onCancel={() => {
          setEditingRow(null);
          editForm.resetFields();
        }}
        onOk={() => editForm.submit()}
        confirmLoading={submitting}
        okText="保存"
        cancelText="取消"
      >
        <Form form={editForm} layout="vertical" onFinish={(values) => void handleSaveInventory(values)}>
          <Form.Item
            label="当前库存"
            name="stock"
            rules={[{ required: true, message: "请输入当前库存" }]}
          >
            <InputNumber min={0} style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item
            label="安全库存"
            name="safeStock"
            rules={[{ required: true, message: "请输入安全库存" }]}
          >
            <InputNumber min={0} style={{ width: "100%" }} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
