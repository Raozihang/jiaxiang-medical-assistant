import {
  AlertOutlined,
  BellOutlined,
  CalendarOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  ImportOutlined,
  LogoutOutlined,
  MedicineBoxOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons";
import type { MenuProps } from "antd";
import { Avatar, Button, Layout, Menu, Space, Tag, Typography } from "antd";
import { useEffect, useMemo, useState } from "react";
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
  icon: React.ReactNode;
  visibleRoles: Array<UserRole | "guest">;
};

const allMenuItems: AppMenuItem[] = [
  { key: "/student/checkin", label: "学生签到", icon: <TeamOutlined />, visibleRoles: ["guest", "doctor", "admin"] },
  { key: "/doctor/visits", label: "就诊队列", icon: <CalendarOutlined />, visibleRoles: ["doctor"] },
  { key: "/doctor/medicines", label: "药品库存", icon: <MedicineBoxOutlined />, visibleRoles: ["doctor"] },
  { key: "/admin/dashboard", label: "管理仪表盘", icon: <DashboardOutlined />, visibleRoles: ["admin"] },
  { key: "/admin/imports", label: "数据导入", icon: <ImportOutlined />, visibleRoles: ["admin"] },
  { key: "/admin/reports", label: "报表统计", icon: <DatabaseOutlined />, visibleRoles: ["admin"] },
  { key: "/admin/notifications", label: "消息通知", icon: <BellOutlined />, visibleRoles: ["admin"] },
  { key: "/admin/safety", label: "安全告警", icon: <AlertOutlined />, visibleRoles: ["admin"] },
];

function getSelectedKey(pathname: string) {
  if (pathname.startsWith("/doctor/visit/")) {
    return "/doctor/visits";
  }
  return pathname;
}

function useCurrentTime() {
  const [time, setTime] = useState(() => new Date());
  useEffect(() => {
    const id = setInterval(() => setTime(new Date()), 30_000);
    return () => clearInterval(id);
  }, []);
  return time.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    weekday: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const currentUser = getStoredUser();
  const currentRole = currentUser?.role ?? "guest";
  const [collapsed, setCollapsed] = useState(false);
  const currentTime = useCurrentTime();

  const selectedKey = useMemo(() => getSelectedKey(location.pathname), [location.pathname]);
  const menuItems = useMemo<MenuProps["items"]>(() => {
    return allMenuItems
      .filter((item) => item.visibleRoles.includes(currentRole))
      .map((item) => ({ key: item.key, label: item.label, icon: item.icon }));
  }, [currentRole]);

  const handleSwitchAccount = () => {
    clearAuth();
    navigate("/login", { replace: true });
  };

  return (
    <Layout className="app-shell">
      <Sider
        theme="dark"
        width={240}
        collapsedWidth={72}
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        trigger={null}
      >
        <div className="sidebar-inner">
          <div className={`app-logo ${collapsed ? "app-logo-collapsed" : ""}`}>
            <MedicineBoxOutlined className="app-logo-icon" />
            {!collapsed && (
              <span className="app-logo-text">
                {env.appTitle}
                <small>Smart Medical</small>
              </span>
            )}
          </div>
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[selectedKey]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
          />
          <div className="sidebar-footer">
            {!collapsed && <span>v0.1.0</span>}
          </div>
        </div>
      </Sider>
      <Layout>
        <Header className="app-header">
          <div className="header-content">
            <div className="header-left">
              <Button
                type="text"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={() => setCollapsed(!collapsed)}
                style={{ fontSize: 16 }}
              />
              <Typography.Title level={4} style={{ margin: 0 }}>
                校园智慧医务控制台
              </Typography.Title>
            </div>
            <div className="header-right">
              <span className="header-time">{currentTime}</span>
              {currentUser ? (
                <Space size={10}>
                  <Avatar
                    size="small"
                    icon={<UserOutlined />}
                    style={{
                      background:
                        currentUser.role === "admin"
                          ? "linear-gradient(135deg,#8b5cf6,#a78bfa)"
                          : "linear-gradient(135deg,#0d9488,#14b8a6)",
                    }}
                  />
                  <Typography.Text strong>{currentUser.name}</Typography.Text>
                  <Tag color={currentUser.role === "admin" ? "purple" : "cyan"}>
                    {currentUser.role === "admin" ? "管理员" : "医生"}
                  </Tag>
                  <Button
                    type="text"
                    size="small"
                    icon={<LogoutOutlined />}
                    onClick={handleSwitchAccount}
                  >
                    退出
                  </Button>
                </Space>
              ) : (
                <Button size="small" type="primary" onClick={() => navigate("/login")}>
                  登录
                </Button>
              )}
            </div>
          </div>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
