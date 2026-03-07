import { http } from "@/shared/api/http";
import type { ApiResponse } from "@/shared/types/api";

export type OverviewReport = {
  today_visits: number;
  observation_students: number;
  stock_warnings: number;
  due_follow_ups: number;
};

export async function getOverviewReport() {
  const response = await http.get<ApiResponse<OverviewReport>>("/reports/overview");
  return response.data.data;
}
