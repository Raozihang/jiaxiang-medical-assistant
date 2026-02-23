import { http } from "@/shared/api/http";
import type { ApiResponse, Paginated } from "@/shared/types/api";

export type Visit = {
  id: string;
  student_id: string;
  student_name: string;
  class_name: string;
  symptoms: string[];
  description: string;
  diagnosis: string;
  prescription: string[];
  destination: string;
  follow_up_at: string | null;
  follow_up_note: string | null;
  created_at: string;
  updated_at: string;
};

export type CreateVisitPayload = {
  student_id: string;
  symptoms: string[];
  description: string;
};

export type UpdateVisitPayload = {
  diagnosis?: string;
  prescription?: string[];
  destination?: string;
  follow_up_at?: string | null;
  follow_up_note?: string | null;
};

export async function listVisits(params: { page?: number; pageSize?: number; studentId?: string }) {
  const response = await http.get<ApiResponse<Paginated<Visit>>>("/visits", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 20,
      student_id: params.studentId,
    },
  });
  return response.data.data;
}

export async function createVisit(payload: CreateVisitPayload) {
  const response = await http.post<ApiResponse<Visit>>("/visits", payload);
  return response.data.data;
}

export async function getVisit(id: string) {
  const response = await http.get<ApiResponse<Visit>>(`/visits/${id}`);
  return response.data.data;
}

export async function updateVisit(id: string, payload: UpdateVisitPayload) {
  const response = await http.patch<ApiResponse<Visit>>(`/visits/${id}`, payload);
  return response.data.data;
}
