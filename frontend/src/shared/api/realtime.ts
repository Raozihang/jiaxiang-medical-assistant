import { getStoredToken } from "@/shared/auth/session";
import { env } from "@/shared/config/env";
import type { Visit } from "./visits";

export type RealtimeMessage<T = unknown> = {
  type: string;
  payload?: T;
  sent_at: string;
};

export type CheckInProgressPayload = {
  session_id: string;
  student_id: string;
  student_name?: string;
  question_id: string;
  question: string;
  symptoms: string[];
  description: string;
  temperature?: number | null;
  temperature_status?: "normal" | "requested" | "timing" | "due" | "measured";
  thermometer_id?: number | null;
  countdown_seconds: number;
  status: "collecting" | "temperature" | "submitted" | "cancelled";
};

export type TemperatureRealtimePayload = {
  session_id: string;
  student_id: string;
  student_name?: string;
  temperature?: number | null;
  temperature_status: "requested" | "timing" | "due" | "measured" | "normal";
  thermometer_id?: number | null;
  countdown_seconds?: number;
};

export type VisitsSnapshotPayload = {
  reason: "connected" | "created" | "updated" | string;
  changed_visit?: Visit;
  items: Visit[];
  page: number;
  page_size: number;
  total: number;
};

function buildRealtimeUrl(path: string, params?: Record<string, string>) {
  const url = new URL(`${env.realtimeBaseUrl}${path}`);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.set(key, value);
    }
  }
  return url.toString();
}

export function connectDoctorRealtime() {
  const token = getStoredToken() ?? "";
  return new WebSocket(buildRealtimeUrl("/realtime/doctor", { token }));
}

export function connectCheckInRealtime() {
  return new WebSocket(buildRealtimeUrl("/realtime/checkin"));
}

export function sendCheckInProgress(socket: WebSocket | null, payload: CheckInProgressPayload) {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    return;
  }
  socket.send(JSON.stringify({ type: "checkin_progress", payload }));
}

export function sendRealtimeMessage(socket: WebSocket | null, type: string, payload: unknown) {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    return;
  }
  socket.send(JSON.stringify({ type, payload }));
}
