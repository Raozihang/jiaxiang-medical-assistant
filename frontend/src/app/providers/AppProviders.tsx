import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import type { PropsWithChildren } from "react";

export function AppProviders({ children }: PropsWithChildren) {
  return <ConfigProvider locale={zhCN}>{children}</ConfigProvider>;
}
