import { Button, Result } from "antd";
import { useNavigate } from "react-router-dom";
import { getStoredUser, resolveHomePath } from "@/shared/auth/session";

export function ForbiddenPage() {
  const navigate = useNavigate();
  const user = getStoredUser();

  return (
    <Result
      status="403"
      title="访问被拒绝"
      subTitle="您没有权限访问此页面。"
      extra={
        <Button type="primary" onClick={() => navigate(resolveHomePath(user?.role), { replace: true })}>
          返回首页
        </Button>
      }
    />
  );
}
