import { http } from "@/shared/api/http";
import { unwrapApiData } from "@/shared/api/helpers";

export type ColumnOption = {
  key: string;
  label: string;
};

export type ReportTemplate = {
  id: string;
  name: string;
  period: string;
  columns: string[];
  title: string;
  created_at: string;
  updated_at: string;
};

export type ReportSchedule = {
  id: string;
  template_id: string;
  cron_expr: string;
  enabled: boolean;
  last_run_at: string | null;
  next_run_at: string | null;
  created_at: string;
  updated_at: string;
};

export type ScheduledReportFile = {
  name: string;
  size_bytes: number;
  modified_at: string;
};

async function downloadBlob(path: string, method: "get" | "post" = "get"): Promise<void> {
  const response = await http.request({
    url: path,
    method,
    responseType: "blob",
    timeout: 30000,
  });

  const disposition = response.headers["content-disposition"] ?? "";
  const match = disposition.match(/filename="?([^"]+)"?/);
  const filename = match?.[1] ?? "报表.xlsx";

  const url = URL.createObjectURL(response.data as Blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = decodeURIComponent(filename);
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export async function getColumnOptions(): Promise<ColumnOption[]> {
  const resp = await http.get("/report-templates/columns");
  return unwrapApiData<ColumnOption[]>(resp.data) ?? [];
}

export async function listTemplates(): Promise<ReportTemplate[]> {
  const resp = await http.get("/report-templates");
  return unwrapApiData<ReportTemplate[]>(resp.data) ?? [];
}

export async function createTemplate(data: {
  name: string;
  period: string;
  columns: string[];
  title?: string;
}): Promise<ReportTemplate> {
  const resp = await http.post("/report-templates", data);
  return unwrapApiData<ReportTemplate>(resp.data)!;
}

export async function updateTemplate(
  id: string,
  data: { name?: string; columns?: string[]; title?: string },
): Promise<ReportTemplate> {
  const resp = await http.patch(`/report-templates/${id}`, data);
  return unwrapApiData<ReportTemplate>(resp.data)!;
}

export async function deleteTemplate(id: string): Promise<void> {
  await http.delete(`/report-templates/${id}`);
}

export async function exportWithTemplate(id: string): Promise<void> {
  await downloadBlob(`/report-templates/${id}/export`);
}

export async function listSchedules(): Promise<ReportSchedule[]> {
  const resp = await http.get("/report-schedules");
  return unwrapApiData<ReportSchedule[]>(resp.data) ?? [];
}

export async function createSchedule(data: {
  template_id: string;
  cron_expr: string;
}): Promise<ReportSchedule> {
  const resp = await http.post("/report-schedules", data);
  return unwrapApiData<ReportSchedule>(resp.data)!;
}

export async function updateSchedule(
  id: string,
  data: { cron_expr?: string; enabled?: boolean },
): Promise<ReportSchedule> {
  const resp = await http.patch(`/report-schedules/${id}`, data);
  return unwrapApiData<ReportSchedule>(resp.data)!;
}

export async function deleteSchedule(id: string): Promise<void> {
  await http.delete(`/report-schedules/${id}`);
}

export async function triggerSchedule(id: string): Promise<void> {
  await downloadBlob(`/report-schedules/${id}/run`, "post");
}

export async function listScheduleFiles(id: string): Promise<ScheduledReportFile[]> {
  const resp = await http.get(`/report-schedules/${id}/files`);
  return unwrapApiData<ScheduledReportFile[]>(resp.data) ?? [];
}

export async function downloadScheduleFile(id: string, fileName: string): Promise<void> {
  await downloadBlob(`/report-schedules/${id}/files/${encodeURIComponent(fileName)}`);
}
