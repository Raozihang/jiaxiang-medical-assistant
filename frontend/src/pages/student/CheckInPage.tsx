import { Button, Card, Form, Input, message, Select, Space, Typography } from "antd";
import { createVisit } from "@/shared/api/visits";

type CheckInForm = {
  studentId: string;
  symptoms: string[];
  description: string;
};

const symptomOptions = [
  { label: "发热", value: "fever" },
  { label: "咳嗽", value: "cough" },
  { label: "头痛", value: "headache" },
  { label: "外伤", value: "injury" },
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
    messageApi.success("签到成功");
    form.resetFields();
  };

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        学生签到
      </Typography.Title>
      <Card>
        <Form layout="vertical" form={form} onFinish={(values) => void handleSubmit(values)}>
          <Form.Item label="学号" name="studentId" rules={[{ required: true, message: "请输入学号" }]}>
            <Input placeholder="请输入学号" />
          </Form.Item>
          <Form.Item label="症状" name="symptoms">
            <Select mode="multiple" options={symptomOptions} placeholder="请选择症状" />
          </Form.Item>
          <Form.Item label="补充描述" name="description">
            <Input.TextArea rows={4} placeholder="额外备注（可选）" />
          </Form.Item>
          <Button htmlType="submit" type="primary">
            提交
          </Button>
        </Form>
      </Card>
    </Space>
  );
}
