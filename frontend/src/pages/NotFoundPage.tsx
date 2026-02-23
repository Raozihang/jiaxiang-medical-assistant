import { Button, Result } from "antd";
import { useNavigate } from "react-router-dom";

export function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <Result
      status="404"
      title="Page not found"
      subTitle="Please check the URL or return to the homepage."
      extra={
        <Button type="primary" onClick={() => navigate("/")}>
          Back Home
        </Button>
      }
    />
  );
}
