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
  { key: "/student/checkin", label: "学生签到", visibleRoles: ["guest", "doctor", "admin"] },
  { key: "/doctor/visits", label: "就诊队列", visibleRoles: ["doctor"] },
  { key: "/doctor/medicines", label: "药品库存", visibleRoles: ["doctor"] },
  { key: "/admin/dashboard", label: "管理仪表盘", visibleRoles: ["admin"] },
  { key: "/admin/imports", label: "数据导入", visibleRoles: ["admin"] },
  { key: "/admin/reports", label: "报表统计", visibleRoles: ["admin"] },
  { key: "/admin/notifications", label: "消息通知", visibleRoles: ["admin"] },
  { key: "/admin/safety", label: "安全告警", visibleRoles: ["admin"] },
];

function getSelectedKey(pathname: string) {
  if (pathname.startsWith("/doctor/visit/")) {
    return "/doctor/visits";
  }
  return pathname;
}

const roleLabels: Record<UserRole, string> = {
  admin: "管理员",
  doctor: "医生",
};

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
          <Space style={{ width: "100%", justifyContent: "space-between" }}>
            <Typography.Title level={4} style={{ margin: 0 }}>
              校园智慧医务控制台
            </Typography.Title>
            {currentUser ? (
              <Space size={12}>
                <Typography.Text>{currentUser.name}</Typography.Text>
                <Tag color={currentUser.role === "admin" ? "purple" : "blue"}>{roleLabels[currentUser.role]}</Tag>
                <Button size="small" onClick={handleLogout}>
                  退出登录
                </Button>
              </Space>
            ) : (
              <Button size="small" type="primary" onClick={() => navigate("/login")}>
                登录
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

