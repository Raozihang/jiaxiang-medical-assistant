import { http } from "@/shared/api/http";
import {
  asRecord,
  normalizePaginated,
  pickFirst,
  toText,
  unwrapApiData,
} from "@/shared/api/helpers";

export type SafetyAlert = {
  id: string;
  level: string;
  type: string;
  title: string;
  description: string;
  source: string;
  status: string;
  created_at: string;
  resolved_at?: string;
};

function toSafetyAlert(item: unknown, index = 0): SafetyAlert {
  const record = asRecord(item);
  const rule = toText(pickFirst(record, ["rule"]), "general");
  return {
    id: toText(pickFirst(record, ["id"]), `alert-${index}`),
    level: rule === "observation_timeout" ? "high" : "medium",
    type: rule,
    title: rule === "observation_timeout" ? "留观超时" : "安全告警",
    description: toText(pickFirst(record, ["message", "description"]), ""),
    source: toText(pickFirst(record, ["student_id", "source"]), "system"),
    status: toText(pickFirst(record, ["status"]), "open"),
    created_at: toText(pickFirst(record, ["triggered_at", "created_at"])),
    resolved_at: toText(pickFirst(record, ["resolved_at"])) || undefined,
  };
}

export async function listSafetyAlerts(params: { page?: number; pageSize?: number; status?: string }) {
  const response = await http.get("/safety/alerts", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
      status: params.status,
    },
  });

  return normalizePaginated(response.data, toSafetyAlert);
}

export async function resolveSafetyAlert(id: string) {
  const response = await http.patch(`/safety/alerts/${id}`, { status: "resolved" });
  return toSafetyAlert(unwrapApiData(response.data));
}
