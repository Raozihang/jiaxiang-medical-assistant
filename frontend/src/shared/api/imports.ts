import { http } from "@/shared/api/http";
import type { ApiResponse, Paginated } from "@/shared/types/api";

export type VisitImportItem = {
  student_id: string;
  symptoms?: string[];
  description?: string;
  diagnosis?: string;
  prescription?: string[];
  destination?: string;
};

export type ImportTaskError = {
  index: number;
  message: string;
};

export type ImportTask = {
  id: string;
  status: string;
  total: number;
  success: number;
  failed: number;
  errors: ImportTaskError[];
  created_at: string;
  updated_at: string;
};

export async function listImportTasks(params: { page?: number; pageSize?: number }) {
  const response = await http.get<ApiResponse<Paginated<ImportTask>>>("/import/tasks", {
    params: { page: params.page ?? 1, page_size: params.pageSize ?? 10 },
  });
  return response.data.data;
}

export async function createImportTask(items: VisitImportItem[]) {
  const response = await http.post<ApiResponse<ImportTask>>("/import/visits", items);
  return response.data.data;
}

export async function getImportTask(id: string) {
  const response = await http.get<ApiResponse<ImportTask>>(`/import/tasks/${id}`);
  return response.data.data;
}
