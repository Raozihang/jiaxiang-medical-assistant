import { http } from "@/shared/api/http";
import {
  asRecord,
  normalizePaginated,
  pickFirst,
  toText,
  unwrapApiData,
} from "@/shared/api/helpers";

export type NotificationChannel = "wechat" | "dingtalk";

export type SendNotificationPayload = {
  channel: NotificationChannel;
  receiver: string;
  message: string;
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

export async function sendNotification(payload: SendNotificationPayload) {
  const response = await http.post("/notifications/send", payload);
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
