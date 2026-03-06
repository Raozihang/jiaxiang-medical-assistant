import {
  asRecord,
  normalizePaginated,
  pickFirst,
  toText,
  unwrapApiData,
} from "@/shared/api/helpers";
import { http } from "@/shared/api/http";

export type NotificationChannel = "wechat" | "dingtalk";

export type SendNotificationPayload = {
  channel: NotificationChannel;
  receiver: string;
  message: string;
};

export type NotificationScenario = "visit_completed" | "observation_notice" | "follow_up_reminder";

export type DispatchScenarioNotificationPayload = {
  scenario: NotificationScenario;
  channel: NotificationChannel;
  receiver: string;
  student_name?: string;
  destination?: string;
  follow_up_at?: string;
  note?: string;
};

export type NotificationLog = {
  id: string;
  channel: string;
  receiver: string;
  message: string;
  status: string;
  error?: string;
  sent_at: string;
};

export type StudentContact = {
  student_id: string;
  student_name: string;
  guardian_name: string;
  guardian_phone: string;
  guardian_relation: string;
};

export type UpdateStudentContactPayload = {
  student_name?: string;
  guardian_name?: string;
  guardian_phone?: string;
  guardian_relation?: string;
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
  request_id: string;
  call_id: string;
  error?: string;
  requested_at: string;
  completed_at?: string;
};

function toNotificationLog(item: unknown, index = 0): NotificationLog {
  const record = asRecord(item);
  return {
    id: toText(pickFirst(record, ["id"]), `log-${index}`),
    channel: toText(pickFirst(record, ["channel"])),
    receiver: toText(pickFirst(record, ["receiver"])),
    message: toText(pickFirst(record, ["message"])),
    status: toText(pickFirst(record, ["status"]), "sent"),
    error: toText(pickFirst(record, ["error"])) || undefined,
    sent_at: toText(pickFirst(record, ["sent_at"])),
  };
}

function toStudentContact(item: unknown, index = 0): StudentContact {
  const record = asRecord(item);
  return {
    student_id: toText(pickFirst(record, ["student_id"]), `student-${index}`),
    student_name: toText(pickFirst(record, ["student_name"])),
    guardian_name: toText(pickFirst(record, ["guardian_name"])),
    guardian_phone: toText(pickFirst(record, ["guardian_phone"])),
    guardian_relation: toText(pickFirst(record, ["guardian_relation"])),
  };
}

function toOutboundCall(item: unknown, index = 0): OutboundCall {
  const record = asRecord(item);
  return {
    id: toText(pickFirst(record, ["id"]), `call-${index}`),
    visit_id: toText(pickFirst(record, ["visit_id"])),
    student_id: toText(pickFirst(record, ["student_id"])),
    student_name: toText(pickFirst(record, ["student_name"])),
    guardian_name: toText(pickFirst(record, ["guardian_name"])),
    guardian_phone: toText(pickFirst(record, ["guardian_phone"])),
    guardian_relation: toText(pickFirst(record, ["guardian_relation"])),
    scenario: toText(pickFirst(record, ["scenario"])),
    provider: toText(pickFirst(record, ["provider"])),
    trigger_source: toText(pickFirst(record, ["trigger_source"])),
    status: toText(pickFirst(record, ["status"])),
    message: toText(pickFirst(record, ["message"])),
    template_code: toText(pickFirst(record, ["template_code"])),
    request_id: toText(pickFirst(record, ["request_id"])),
    call_id: toText(pickFirst(record, ["call_id"])),
    error: toText(pickFirst(record, ["error"])) || undefined,
    requested_at: toText(pickFirst(record, ["requested_at"])),
    completed_at: toText(pickFirst(record, ["completed_at"])) || undefined,
  };
}

export async function sendNotification(payload: SendNotificationPayload) {
  const response = await http.post("/notifications/send", payload);
  return toNotificationLog(unwrapApiData(response.data));
}

export async function dispatchScenarioNotification(payload: DispatchScenarioNotificationPayload) {
  const response = await http.post("/notifications/dispatch", payload);
  return toNotificationLog(unwrapApiData(response.data));
}

export async function listNotificationLogs(params: { page?: number; pageSize?: number }) {
  const response = await http.get("/notifications/logs", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
    },
  });

  return normalizePaginated(response.data, toNotificationLog);
}

export async function listStudentContacts(params: {
  page?: number;
  pageSize?: number;
  keyword?: string;
}) {
  const response = await http.get("/students/contacts", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
      keyword: params.keyword ?? undefined,
    },
  });

  return normalizePaginated(response.data, toStudentContact);
}

export async function updateStudentContact(
  studentId: string,
  payload: UpdateStudentContactPayload,
) {
  const response = await http.put(`/students/${studentId}/contact`, payload);
  return toStudentContact(unwrapApiData(response.data));
}

export async function listOutboundCalls(params: {
  page?: number;
  pageSize?: number;
  status?: string;
  studentId?: string;
  keyword?: string;
}) {
  const response = await http.get("/outbound-calls", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
      status: params.status ?? undefined,
      student_id: params.studentId ?? undefined,
      keyword: params.keyword ?? undefined,
    },
  });

  return normalizePaginated(response.data, toOutboundCall);
}

export async function retryOutboundCall(id: string) {
  const response = await http.post(`/outbound-calls/${id}/retry`);
  return toOutboundCall(unwrapApiData(response.data));
}
