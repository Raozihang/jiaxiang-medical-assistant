import { Button, Card, Form, Input, message, Select, Space, Typography } from "antd";
import { createVisit } from "@/shared/api/visits";

type CheckInForm = {
  studentId: string;
  symptoms: string[];
  description: string;
};

const symptomOptions = [
  { label: "Fever", value: "fever" },
  { label: "Cough", value: "cough" },
  { label: "Headache", value: "headache" },
  { label: "Injury", value: "injury" },
];

export function CheckInPage() {
  const [form] = Form.useForm<CheckInForm>();
  const [messageApi, contextHolder] = message.useMessage();

  const handleSubmit = async (values: CheckInForm) => {
    await createVisit({
      student_id: values.studentId,
      symptoms: values.symptoms ?? [],
      description: values.description ?? "",
    });
    messageApi.success("Check-in created");
    form.resetFields();
  };

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        Student Check-In
      </Typography.Title>
      <Card>
        <Form layout="vertical" form={form} onFinish={(values) => void handleSubmit(values)}>
          <Form.Item label="Student ID" name="studentId" rules={[{ required: true }]}>
            <Input placeholder="Enter student ID" />
          </Form.Item>
          <Form.Item label="Symptoms" name="symptoms">
            <Select mode="multiple" options={symptomOptions} placeholder="Select symptoms" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea rows={4} placeholder="Additional notes (optional)" />
          </Form.Item>
          <Button htmlType="submit" type="primary">
            Submit
          </Button>
        </Form>
      </Card>
    </Space>
  );
}
