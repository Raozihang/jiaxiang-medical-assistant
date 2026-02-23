import type { MenuProps } from "antd";
import { Button, Layout, Menu, Space, Tag, Typography } from "antd";
import { useMemo } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  clearAuth,
  getStoredUser,
  type UserRole,
} from "@/shared/auth/session";
import { env } from "@/shared/config/env";

const { Header, Content, Sider } = Layout;

type AppMenuItem = {
  key: string;
  label: string;
  visibleRoles: Array<UserRole | "guest">;
};

const allMenuItems: AppMenuItem[] = [
  { key: "/student/checkin", label: "Student Check-In", visibleRoles: ["guest", "doctor", "admin"] },
  { key: "/doctor/visits", label: "Visit Queue", visibleRoles: ["doctor"] },
  { key: "/doctor/medicines", label: "Medicine Inventory", visibleRoles: ["doctor"] },
  { key: "/admin/dashboard", label: "Dashboard", visibleRoles: ["admin"] },
  { key: "/admin/imports", label: "Data Imports", visibleRoles: ["admin"] },
  { key: "/admin/reports", label: "Reports", visibleRoles: ["admin"] },
  { key: "/admin/notifications", label: "Notifications", visibleRoles: ["admin"] },
  { key: "/admin/safety", label: "Safety Alerts", visibleRoles: ["admin"] },
];

function getSelectedKey(pathname: string) {
  if (pathname.startsWith("/doctor/visit/")) {
    return "/doctor/visits";
  }
  return pathname;
}

export function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const currentUser = getStoredUser();
  const currentRole = currentUser?.role ?? "guest";

  const selectedKey = useMemo(() => getSelectedKey(location.pathname), [location.pathname]);
  const menuItems = useMemo<MenuProps["items"]>(() => {
    return allMenuItems
      .filter((item) => item.visibleRoles.includes(currentRole))
      .map((item) => ({ key: item.key, label: item.label }));
  }, [currentRole]);

  const handleSwitchAccount = () => {
    clearAuth();
    navigate("/login", { replace: true });
  };

  return (
    <Layout className="app-shell">
      <Sider theme="light" width={240}>
        <div className="app-logo">{env.appTitle}</div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <Space style={{ width: "100%", justifyContent: "space-between" }}>
            <Typography.Title level={4} style={{ margin: 0 }}>
              Campus Medical Console
            </Typography.Title>
            {currentUser ? (
              <Space size={12}>
                <Typography.Text>{currentUser.name}</Typography.Text>
                <Tag color={currentUser.role === "admin" ? "purple" : "blue"}>{currentUser.role}</Tag>
                <Button size="small" onClick={handleSwitchAccount}>
                  Switch Account
                </Button>
              </Space>
            ) : (
              <Button size="small" type="primary" onClick={() => navigate("/login")}>
                Login
              </Button>
            )}
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}

