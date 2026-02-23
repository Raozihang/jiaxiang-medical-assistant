import { Button, Result } from "antd";
import { useNavigate } from "react-router-dom";
import { getStoredUser, resolveHomePath } from "@/shared/auth/session";

export function ForbiddenPage() {
  const navigate = useNavigate();
  const user = getStoredUser();

  return (
    <Result
      status="403"
      title="Access denied"
      subTitle="You do not have permission to access this page."
      extra={
        <Button type="primary" onClick={() => navigate(resolveHomePath(user?.role), { replace: true })}>
          Back to home
        </Button>
      }
    />
  );
}
