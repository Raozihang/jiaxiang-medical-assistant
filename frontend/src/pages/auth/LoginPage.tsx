import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { Button, Card, Form, Input, Space, Typography, message } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { login } from "@/shared/api/auth";
import { resolveHomePath } from "@/shared/auth/session";
import { env } from "@/shared/config/env";

type LoginForm = {
  account: string;
  password: string;
};

function getErrorMessage(error: unknown, fallback: string) {
  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof error.response === "object" &&
    error.response !== null &&
    "data" in error.response &&
    typeof error.response.data === "object" &&
    error.response.data !== null &&
    "message" in error.response.data &&
    typeof error.response.data.message === "string"
  ) {
    return error.response.data.message;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}

export function LoginPage() {
  const navigate = useNavigate();
  const [submitting, setSubmitting] = useState(false);
  const [messageApi, contextHolder] = message.useMessage();

  const handleFinish = async (values: LoginForm) => {
    setSubmitting(true);
    try {
      const result = await login(values.account, values.password);
      messageApi.success("登录成功");
      navigate(resolveHomePath(result.user.role), { replace: true });
    } catch (error) {
      messageApi.error(getErrorMessage(error, "登录失败，请检查账号或密码"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        background: "linear-gradient(180deg, #f8fafc 0%, #e2e8f0 100%)",
        padding: 24,
      }}
    >
      {contextHolder}
      <Card style={{ width: 420, maxWidth: "100%", borderRadius: 12 }}>
        <Space direction="vertical" size={24} style={{ width: "100%" }}>
          <div style={{ textAlign: "center" }}>
            <Typography.Title level={3} style={{ marginBottom: 4 }}>
              欢迎登录
            </Typography.Title>
            <Typography.Text type="secondary">{env.appTitle}</Typography.Text>
          </div>
          <Form<LoginForm> layout="vertical" onFinish={(values) => void handleFinish(values)}>
            <Form.Item label="账号" name="account" rules={[{ required: true, message: "请输入账号" }]}>
              <Input
                prefix={<UserOutlined style={{ color: "rgba(0,0,0,0.25)" }} />}
                placeholder="doctor / admin"
                autoComplete="username"
                size="large"
              />
            </Form.Item>
            <Form.Item label="密码" name="password" rules={[{ required: true, message: "请输入密码" }]}>
              <Input.Password
                prefix={<LockOutlined style={{ color: "rgba(0,0,0,0.25)" }} />}
                placeholder="请输入密码"
                autoComplete="current-password"
                size="large"
              />
            </Form.Item>
            <Button type="primary" htmlType="submit" block size="large" loading={submitting}>
              登录
            </Button>
          </Form>
        </Space>
      </Card>
    </div>
  );
}
