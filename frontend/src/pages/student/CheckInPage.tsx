import { CheckCircleOutlined, MedicineBoxOutlined } from "@ant-design/icons";
import { Button, Card, Form, Input, message, Result, Select, Space, Typography } from "antd";
import { useState } from "react";
import { createVisit } from "@/shared/api/visits";

type CheckInForm = {
  studentId: string;
  symptoms: string[];
  description: string;
};

const symptomOptions = [
  { label: "🌡️ 发热", value: "fever" },
  { label: "😷 咳嗽", value: "cough" },
  { label: "🤕 头痛", value: "headache" },
  { label: "🩹 外伤", value: "injury" },
];

export function CheckInPage() {
  const [form] = Form.useForm<CheckInForm>();
  const [messageApi, contextHolder] = message.useMessage();
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (values: CheckInForm) => {
    await createVisit({
      student_id: values.studentId,
      symptoms: values.symptoms ?? [],
      description: values.description ?? "",
    });
    messageApi.success("签到成功");
    setSubmitted(true);
  };

  const handleReset = () => {
    form.resetFields();
    setSubmitted(false);
  };

  if (submitted) {
    return (
      <div className="student-checkin-wrapper">
        {contextHolder}
        <Result
          icon={<CheckCircleOutlined style={{ color: "#0d9488" }} />}
          title="签到成功！"
          subTitle="请在等候区稍候，医生将会尽快接诊。"
          extra={
            <Button type="primary" size="large" onClick={handleReset}>
              继续签到
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div className="student-checkin-wrapper">
      {contextHolder}
      <Space direction="vertical" size={20} style={{ width: "100%" }}>
        <div style={{ textAlign: "center" }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 16,
              background: "linear-gradient(135deg, #0d9488, #14b8a6)",
              color: "#fff",
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              fontSize: 28,
              marginBottom: 12,
            }}
          >
            <MedicineBoxOutlined />
          </div>
          <Typography.Title level={3} style={{ marginBottom: 4 }}>
            学生签到
          </Typography.Title>
          <Typography.Text type="secondary">
            请填写您的信息，医生将尽快为您诊疗
          </Typography.Text>
        </div>
        <Card>
          <Form layout="vertical" form={form} onFinish={(values) => void handleSubmit(values)}>
            <Form.Item label="学号" name="studentId" rules={[{ required: true, message: "请输入学号" }]}>
              <Input placeholder="请输入学号" size="large" />
            </Form.Item>
            <Form.Item label="症状" name="symptoms">
              <Select mode="multiple" options={symptomOptions} placeholder="请选择症状" size="large" />
            </Form.Item>
            <Form.Item label="补充描述" name="description">
              <Input.TextArea rows={4} placeholder="额外备注（可选）" />
            </Form.Item>
            <Button htmlType="submit" type="primary" size="large" block style={{ height: 44, fontWeight: 600 }}>
              提交签到
            </Button>
          </Form>
        </Card>
      </Space>
    </div>
  );
}
