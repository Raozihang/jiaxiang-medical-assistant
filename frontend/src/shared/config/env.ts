const defaultApiBaseUrl = "http://localhost:8080/api/v1";

function resolveRealtimeBaseUrl(apiBaseUrl: string) {
  const configured = import.meta.env.VITE_REALTIME_BASE_URL;
  if (configured) {
    return configured;
  }

  try {
    const url = new URL(apiBaseUrl);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    return url.toString().replace(/\/$/, "");
  } catch {
    return "ws://localhost:8080/api/v1";
  }
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? defaultApiBaseUrl;

export const env = {
  appTitle: import.meta.env.VITE_APP_TITLE ?? "嘉祥智能医务室助手",
  apiBaseUrl,
  realtimeBaseUrl: resolveRealtimeBaseUrl(apiBaseUrl),
};
