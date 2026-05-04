import type { MenuProps } from "antd";
import { Button, Layout, Menu, Modal, notification, Space, Tag, Typography } from "antd";
import { useEffect, useMemo, useRef } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { listSafetyAlerts, type SafetyAlert } from "@/shared/api/safety";
import { clearAuth, getStoredUser, type UserRole } from "@/shared/auth/session";
import { env } from "@/shared/config/env";

const { Header, Content, Sider } = Layout;

type AppMenuItem = {
  key: string;
  label: string;
  visibleRoles: Array<UserRole | "guest">;
};

const allMenuItems: AppMenuItem[] = [
  {
    key: "/student/checkin",
    label: "学生签到",
    visibleRoles: ["guest", "student", "doctor", "admin"],
  },
  { key: "/doctor/visits", label: "就诊队列", visibleRoles: ["doctor", "admin"] },
  { key: "/doctor/medicines", label: "药品库存", visibleRoles: ["doctor", "admin"] },
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
  student: "学生",
  admin: "管理员",
  doctor: "医生",
};

const roleTagColors: Record<UserRole, string> = {
  student: "green",
  doctor: "blue",
  admin: "purple",
};

function safetyLevelColor(level: string) {
  if (["critical", "high", "urgent"].includes(level)) {
    return "red";
  }
  if (["medium", "warning"].includes(level)) {
    return "orange";
  }
  return "blue";
}

function isHighPriorityAlert(alert: SafetyAlert) {
  return ["critical", "high", "urgent"].includes(alert.level);
}

function formatAlertStudent(alert: SafetyAlert) {
  const studentName = alert.student_name?.trim();
  const studentId = (alert.student_id || alert.source || "").trim();
  if (studentName && studentId && studentName !== studentId) {
    return `${studentName}（${studentId}）`;
  }
  return studentName || studentId || "未识别学生";
}

function alertDescriptionWithStudent(alert: SafetyAlert) {
  const description = alert.description || "有新的未处理安全告警，请及时查看。";
  return `学生：${formatAlertStudent(alert)}。${description}`;
}

export function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const currentUser = getStoredUser();
  const currentRole = currentUser?.role ?? "guest";
  const [notificationApi, notificationHolder] = notification.useNotification();
  const [modal, modalHolder] = Modal.useModal();
  const remindedAlertIdsRef = useRef<Set<string>>(new Set());
  const alertModalOpenRef = useRef(false);

  const selectedKey = useMemo(() => getSelectedKey(location.pathname), [location.pathname]);
  const menuItems = useMemo<MenuProps["items"]>(() => {
    return allMenuItems
      .filter((item) => item.visibleRoles.includes(currentRole))
      .map((item) => ({ key: item.key, label: item.label }));
  }, [currentRole]);

  useEffect(() => {
    if (currentRole !== "admin") {
      return;
    }

    let disposed = false;

    const remindOpenAlerts = async () => {
      try {
        const data = await listSafetyAlerts({ page: 1, pageSize: 5, status: "open" });
        if (disposed || !data.items.length) {
          return;
        }

        const newAlerts = data.items.filter((alert) => !remindedAlertIdsRef.current.has(alert.id));
        for (const alert of newAlerts) {
          remindedAlertIdsRef.current.add(alert.id);
          notificationApi.warning({
            key: `safety-${alert.id}`,
            message: alert.title || "安全告警",
            description: alertDescriptionWithStudent(alert),
            duration: 0,
            btn: (
              <Button size="small" type="primary" onClick={() => navigate("/admin/safety")}>
                查看告警
              </Button>
            ),
          });
        }

        const priorityAlert = newAlerts.find(isHighPriorityAlert);
        if (!priorityAlert || alertModalOpenRef.current) {
          return;
        }

        alertModalOpenRef.current = true;
        modal.warning({
          title: `高优先级告警：${priorityAlert.title}`,
          content: (
            <Space direction="vertical" size={8}>
              <Tag color={safetyLevelColor(priorityAlert.level)}>{priorityAlert.level}</Tag>
              <Typography.Text strong>学生：{formatAlertStudent(priorityAlert)}</Typography.Text>
              <Typography.Text>
                {priorityAlert.description || "请尽快处理该告警。"}
              </Typography.Text>
            </Space>
          ),
          okText: "去处理",
          cancelText: "稍后",
          onOk: () => navigate("/admin/safety"),
          afterClose: () => {
            alertModalOpenRef.current = false;
          },
        });
      } catch {
        // 告警提醒不阻塞主界面；安全告警页仍可手动刷新查看。
      }
    };

    void remindOpenAlerts();
    const timer = window.setInterval(() => void remindOpenAlerts(), 60000);

    return () => {
      disposed = true;
      window.clearInterval(timer);
    };
  }, [currentRole, modal, navigate, notificationApi]);

  const handleLogout = () => {
    clearAuth();
    navigate("/login", { replace: true });
  };

  return (
    <>
      {notificationHolder}
      {modalHolder}
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
                  <Tag color={roleTagColors[currentUser.role]}>{roleLabels[currentUser.role]}</Tag>
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
    </>
  );
}
