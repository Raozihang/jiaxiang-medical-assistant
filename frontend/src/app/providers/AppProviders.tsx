import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import type { PropsWithChildren } from "react";
import { useEffect } from "react";
import { login } from "@/shared/api/auth";
import { getStoredToken } from "@/shared/api/http";

export function AppProviders({ children }: PropsWithChildren) {
  useEffect(() => {
    if (!getStoredToken()) {
      void login();
    }
  }, []);

  return <ConfigProvider locale={zhCN}>{children}</ConfigProvider>;
}
