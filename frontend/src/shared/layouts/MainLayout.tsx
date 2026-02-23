import type { MenuProps } from "antd";
import { Layout, Menu, Typography } from "antd";
import { useMemo } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { env } from "@/shared/config/env";

const { Header, Content, Sider } = Layout;

const menuItems: MenuProps["items"] = [
  { key: "/student/checkin", label: "Student Check-In" },
  { key: "/doctor/visits", label: "Visit Queue" },
  { key: "/doctor/medicines", label: "Medicine Inventory" },
  { key: "/admin/dashboard", label: "Dashboard" },
  { key: "/admin/imports", label: "Data Imports" },
  { key: "/admin/reports", label: "Reports" },
  { key: "/admin/notifications", label: "Notifications" },
  { key: "/admin/safety", label: "Safety Alerts" },
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

  const selectedKey = useMemo(() => getSelectedKey(location.pathname), [location.pathname]);

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
          <Typography.Title level={4} style={{ margin: 0 }}>
            Campus Medical Console
          </Typography.Title>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}

