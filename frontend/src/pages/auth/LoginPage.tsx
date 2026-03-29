import { LockOutlined, MedicineBoxOutlined, UserOutlined } from "@ant-design/icons";
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
      <div className="auth-bubbles">
        <div className="auth-bubble" />
        <div className="auth-bubble" />
        <div className="auth-bubble" />
      </div>
      {contextHolder}
      <Card className="auth-card">
        <Space direction="vertical" size={24} style={{ width: "100%" }}>
          <div style={{ textAlign: "center" }}>
            <div className="auth-logo" style={{ margin: "0 auto 16px" }}>
              <MedicineBoxOutlined />
            </div>
            <Typography.Title level={3} style={{ marginBottom: 4 }}>
              欢迎使用
            </Typography.Title>
            <Typography.Text type="secondary" style={{ fontSize: 15 }}>
              {env.appTitle}
            </Typography.Text>
          </div>
          <Form<LoginForm> layout="vertical" onFinish={(values) => void handleFinish(values)}>
            <Form.Item
              label="账号"
              name="account"
              rules={[{ required: true, message: "请输入账号" }]}
            >
              <Input
                prefix={<UserOutlined style={{ color: "rgba(0,0,0,0.25)" }} />}
                placeholder="doctor / admin"
                autoComplete="username"
                size="large"
              />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: "请输入密码" }]}
            >
              <Input.Password
                prefix={<LockOutlined style={{ color: "rgba(0,0,0,0.25)" }} />}
                placeholder="请输入密码"
                autoComplete="current-password"
                size="large"
              />
            </Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              block
              size="large"
              loading={submitting}
              style={{ marginTop: 8, height: 44, fontWeight: 600 }}
            >
              登录
            </Button>
          </Form>
          <div style={{ textAlign: "center" }}>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              v0.1.0 · 仅限授权人员使用
            </Typography.Text>
          </div>
        </Space>
      </Card>
    </div>
  );
}
