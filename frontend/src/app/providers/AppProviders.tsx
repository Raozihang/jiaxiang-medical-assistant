import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import type { PropsWithChildren } from "react";
import { useEffect, useState } from "react";
import { login } from "@/shared/api/auth";
import { getStoredToken } from "@/shared/api/http";

export function AppProviders({ children }: PropsWithChildren) {
  const [authReady, setAuthReady] = useState(false);

  useEffect(() => {
    let canceled = false;

    const initializeAuth = async () => {
      if (!getStoredToken()) {
        try {
          await login();
        } catch (error) {
          console.error("auto login failed", error);
        }
      }

      if (!canceled) {
        setAuthReady(true);
      }
    };

    void initializeAuth();

    return () => {
      canceled = true;
    };
  }, []);

  if (!authReady) {
    return (
      <ConfigProvider locale={zhCN}>
        <div className="app-loading">Loading...</div>
      </ConfigProvider>
    );
  }

  return <ConfigProvider locale={zhCN}>{children}</ConfigProvider>;
}
