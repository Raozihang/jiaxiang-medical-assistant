import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import type { PropsWithChildren } from "react";

const theme = {
  token: {
    colorPrimary: "#0d9488",
    colorSuccess: "#10b981",
    colorWarning: "#f59e0b",
    colorError: "#ef4444",
    colorInfo: "#0ea5e9",
    borderRadius: 10,
    fontFamily:
      "'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Arial, sans-serif",
    fontSize: 14,
    colorBgLayout: "#f0f5f9",
    controlHeight: 38,
  },
  components: {
    Layout: {
      siderBg: "#0f172a",
      headerBg: "rgba(255,255,255,0.85)",
      bodyBg: "#f0f5f9",
    },
    Menu: {
      darkItemBg: "transparent",
      darkItemColor: "rgba(255,255,255,0.65)",
      darkItemHoverColor: "#ffffff",
      darkItemSelectedBg: "rgba(13,148,136,0.25)",
      darkItemSelectedColor: "#5eead4",
      darkSubMenuItemBg: "transparent",
      itemBorderRadius: 8,
      itemMarginInline: 8,
    },
    Card: {
      borderRadiusLG: 14,
    },
    Button: {
      borderRadius: 8,
    },
    Table: {
      borderRadius: 10,
      headerBg: "#f0f5f9",
    },
  },
} as const;

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <ConfigProvider locale={zhCN} theme={theme}>
      {children}
    </ConfigProvider>
  );
}
