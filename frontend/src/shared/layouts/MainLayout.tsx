import type { MenuProps } from "antd";
import { Button, Layout, Menu, Space, Tag, Typography } from "antd";
import { useMemo } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { clearAuth, getStoredUser, type UserRole } from "@/shared/auth/session";
import { env } from "@/shared/config/env";

const { Header, Content, Sider } = Layout;

type MenuRole = UserRole | "guest";

type AppMenuItem = {
  key: string;
  label: string;
  visibleRoles: MenuRole[];
};

const allMenuItems: AppMenuItem[] = [
  { key: "/student/checkin", label: "Student Check-In", visibleRoles: ["guest", "doctor", "admin"] },
  { key: "/doctor/visits", label: "Visit Queue", visibleRoles: ["doctor"] },
  { key: "/doctor/medicines", label: "Medicine Inventory", visibleRoles: ["doctor"] },
  { key: "/admin/dashboard", label: "Dashboard", visibleRoles: ["admin"] },
  { key: "/admin/notifications", label: "Smart Outbound", visibleRoles: ["admin"] },
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
  const currentRole: MenuRole = currentUser?.role ?? "guest";

  const selectedKey = useMemo(() => getSelectedKey(location.pathname), [location.pathname]);
  const menuItems = useMemo<MenuProps["items"]>(() => {
    return allMenuItems
      .filter((item) => item.visibleRoles.includes(currentRole))
      .map((item) => ({ key: item.key, label: item.label }));
  }, [currentRole]);

  const handleLogout = () => {
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
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12 }}>
            <Typography.Title level={4} style={{ margin: 0 }}>
              Campus Medical Console
            </Typography.Title>
            {currentUser ? (
              <Space size={12}>
                <Typography.Text>{currentUser.name}</Typography.Text>
                <Tag color={currentUser.role === "admin" ? "purple" : "blue"}>
                  {currentUser.role === "admin" ? "管理员" : "医生"}
                </Tag>
                <Button size="small" onClick={handleLogout}>
                  退出
                </Button>
              </Space>
            ) : (
              <Button size="small" type="primary" onClick={() => navigate("/login")}>
                登录
              </Button>
            )}
          </div>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
