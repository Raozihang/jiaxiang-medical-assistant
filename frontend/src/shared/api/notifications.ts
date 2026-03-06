import { http } from "@/shared/api/http";
import type { ApiResponse, Paginated } from "@/shared/types/api";

export type StudentContact = {
  student_id: string;
  student_name: string;
  guardian_name: string;
  guardian_phone: string;
  guardian_relation: string;
};

export type OutboundCall = {
  id: string;
  visit_id: string;
  student_id: string;
  student_name: string;
  guardian_name: string;
  guardian_phone: string;
  guardian_relation: string;
  scenario: string;
  provider: string;
  trigger_source: string;
  status: string;
  message: string;
  template_code: string;
  template_params: string;
  request_id: string;
  call_id: string;
  error?: string;
  response_raw?: string;
  retry_of_id?: string;
  requested_at: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

type ListParams = {
  page?: number;
  pageSize?: number;
  keyword?: string;
  status?: string;
  studentId?: string;
};

type UpdateStudentContactPayload = {
  student_name?: string;
  guardian_name?: string;
  guardian_phone?: string;
  guardian_relation?: string;
};

export async function listStudentContacts(params: ListParams = {}) {
  const response = await http.get<ApiResponse<Paginated<StudentContact>>>("/students/contacts", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
      keyword: params.keyword,
    },
  });
  return response.data.data;
}

export async function updateStudentContact(studentId: string, payload: UpdateStudentContactPayload) {
  const response = await http.put<ApiResponse<StudentContact>>(`/students/${studentId}/contact`, payload);
  return response.data.data;
}

export async function listOutboundCalls(params: ListParams = {}) {
  const response = await http.get<ApiResponse<Paginated<OutboundCall>>>("/outbound-calls", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
      keyword: params.keyword,
      status: params.status,
      student_id: params.studentId,
    },
  });
  return response.data.data;
}

export async function retryOutboundCall(id: string) {
  const response = await http.post<ApiResponse<OutboundCall>>(`/outbound-calls/${id}/retry`);
  return response.data.data;
}
