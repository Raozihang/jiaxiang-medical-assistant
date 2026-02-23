import { Button, Card, Form, Input, Space, Typography, message } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { login } from "@/shared/api/auth";
import { getErrorMessage } from "@/shared/api/helpers";
import { resolveHomePath } from "@/shared/auth/session";
import { env } from "@/shared/config/env";

type LoginForm = {
  account: string;
  password: string;
};

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
    <div className="auth-page">
      {contextHolder}
      <Card className="auth-card">
        <Space direction="vertical" size={20} style={{ width: "100%" }}>
          <div>
            <Typography.Title level={3} style={{ marginBottom: 8 }}>
              欢迎使用
            </Typography.Title>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {env.appTitle}
            </Typography.Paragraph>
          </div>
          <Form<LoginForm> layout="vertical" onFinish={(values) => void handleFinish(values)}>
            <Form.Item
              label="账号"
              name="account"
              rules={[{ required: true, message: "请输入账号" }]}
            >
              <Input placeholder="doctor / admin" autoComplete="username" />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: "请输入密码" }]}
            >
              <Input.Password placeholder="请输入密码" autoComplete="current-password" />
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={submitting}>
              登录
            </Button>
          </Form>
        </Space>
      </Card>
    </div>
  );
}
