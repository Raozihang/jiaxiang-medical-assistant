import { Button, Card, Form, Input, message, Select, Space, Tag, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getVisit, updateVisit, type Visit } from "@/shared/api/visits";

type UpdateForm = {
  diagnosis: string;
  prescription: string;
  destination: string;
};

export function VisitDetailPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [visit, setVisit] = useState<Visit | null>(null);
  const [form] = Form.useForm<UpdateForm>();
  const [messageApi, contextHolder] = message.useMessage();

  const loadDetail = useCallback(async () => {
    if (!id) {
      return;
    }

    setLoading(true);
    try {
      const data = await getVisit(id);
      setVisit(data);
      form.setFieldsValue({
        diagnosis: data.diagnosis ?? "",
        prescription: data.prescription.join(", "),
        destination: data.destination || "observation",
      });
    } finally {
      setLoading(false);
    }
  }, [form, id]);

  useEffect(() => {
    void loadDetail();
  }, [loadDetail]);

  const handleSave = async (values: UpdateForm) => {
    if (!id) {
      return;
    }

    const prescription = values.prescription
      .split(",")
      .map((item) => item.trim())
      .filter((item) => item.length > 0);

    await updateVisit(id, {
      diagnosis: values.diagnosis,
      prescription,
      destination: values.destination,
    });
    messageApi.success("Visit updated");
    await loadDetail();
  };

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        Visit Detail
      </Typography.Title>
      <Card loading={loading}>
        {visit ? (
          <>
            <Typography.Paragraph>
              <strong>ID:</strong> {visit.id}
            </Typography.Paragraph>
            <Typography.Paragraph>
              <strong>Student:</strong> {visit.student_name} / {visit.class_name}
            </Typography.Paragraph>
            <Typography.Paragraph>
              <strong>Symptoms:</strong> {visit.symptoms.join(", ") || "-"}
            </Typography.Paragraph>
            <Typography.Paragraph>
              <strong>Description:</strong> {visit.description || "-"}
            </Typography.Paragraph>
            <Typography.Paragraph>
              <strong>Current Destination:</strong>{" "}
              <Tag color={visit.destination === "urgent" ? "red" : "blue"}>{visit.destination}</Tag>
            </Typography.Paragraph>
            <Form layout="vertical" form={form} onFinish={(values) => void handleSave(values)}>
              <Form.Item label="Diagnosis" name="diagnosis">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item label="Prescription (comma separated)" name="prescription">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item label="Destination" name="destination">
                <Select
                  options={[
                    { label: "Observation", value: "observation" },
                    { label: "Return Class", value: "return_class" },
                    { label: "Hospital Referral", value: "hospital" },
                    { label: "Urgent", value: "urgent" },
                  ]}
                />
              </Form.Item>
              <Button htmlType="submit" type="primary">
                Save
              </Button>
            </Form>
          </>
        ) : (
          <Typography.Text type="secondary">No data</Typography.Text>
        )}
      </Card>
    </Space>
  );
}
