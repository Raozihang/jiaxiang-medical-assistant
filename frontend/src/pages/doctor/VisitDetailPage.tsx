import { Button, Card, Col, DatePicker, Form, Input, Row, Select, Space, Tag, Typography, message } from "antd";
import dayjs, { type Dayjs } from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getVisit, updateVisit, type Visit } from "@/shared/api/visits";

type UpdateForm = {
  diagnosis: string;
  prescription: string;
  destination: string;
  follow_up_at: Dayjs | null;
  follow_up_note: string;
};

function formatFollowUpAt(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const parsed = dayjs(value);
  if (!parsed.isValid()) {
    return "-";
  }

  return parsed.format("YYYY-MM-DD HH:mm");
}

function parseFollowUpAt(value: string | null | undefined) {
  if (!value) {
    return null;
  }

  const parsed = dayjs(value);
  return parsed.isValid() ? parsed : null;
}

function destinationTagColor(destination: string) {
  switch (destination) {
    case "urgent":
      return "red";
    case "leave_school":
    case "hospital":
    case "referred":
      return "orange";
    case "return_class":
      return "green";
    default:
      return "blue";
  }
}

export function VisitDetailPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
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
        follow_up_at: parseFollowUpAt(data.follow_up_at),
        follow_up_note: data.follow_up_note ?? "",
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

    setSaving(true);
    try {
      await updateVisit(id, {
        diagnosis: values.diagnosis,
        prescription,
        destination: values.destination,
        follow_up_at: values.follow_up_at ? values.follow_up_at.toDate().toISOString() : "",
        follow_up_note: values.follow_up_note.trim(),
      });
      messageApi.success("Visit updated");
      await loadDetail();
    } finally {
      setSaving(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        Visit Detail
      </Typography.Title>
      <Card loading={loading}>
        {visit ? (
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Typography.Paragraph>
                  <strong>ID:</strong> {visit.id}
                </Typography.Paragraph>
                <Typography.Paragraph>
                  <strong>Student:</strong> {visit.student_name} / {visit.class_name}
                </Typography.Paragraph>
                <Typography.Paragraph>
                  <strong>Symptoms:</strong> {visit.symptoms.join(", ") || "-"}
                </Typography.Paragraph>
              </Col>
              <Col span={12}>
                <Typography.Paragraph>
                  <strong>Description:</strong> {visit.description || "-"}
                </Typography.Paragraph>
                <Typography.Paragraph>
                  <strong>Current Destination:</strong>{" "}
                  <Tag color={destinationTagColor(visit.destination)}>{visit.destination}</Tag>
                </Typography.Paragraph>
                <Typography.Paragraph>
                  <strong>Follow-up Time:</strong> {formatFollowUpAt(visit.follow_up_at)}
                </Typography.Paragraph>
                <Typography.Paragraph>
                  <strong>Follow-up Note:</strong> {visit.follow_up_note?.trim() || "-"}
                </Typography.Paragraph>
              </Col>
            </Row>
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
                    { label: "Leave School", value: "leave_school" },
                    { label: "Hospital Referral", value: "hospital" },
                    { label: "Referred", value: "referred" },
                    { label: "Urgent", value: "urgent" },
                  ]}
                />
              </Form.Item>
              <Form.Item label="Follow-up Time" name="follow_up_at">
                <DatePicker showTime format="YYYY-MM-DD HH:mm" style={{ width: "100%" }} allowClear />
              </Form.Item>
              <Form.Item label="Follow-up Note" name="follow_up_note">
                <Input.TextArea rows={3} placeholder="Optional reminder note" />
              </Form.Item>
              <Button htmlType="submit" type="primary" loading={saving}>
                Save
              </Button>
            </Form>
          </Space>
        ) : (
          <Typography.Text type="secondary">No data</Typography.Text>
        )}
      </Card>
    </Space>
  );
}

