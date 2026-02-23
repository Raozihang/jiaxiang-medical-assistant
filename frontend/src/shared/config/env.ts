const defaultApiBaseUrl = "http://localhost:8080/api/v1";

export const env = {
  appTitle: import.meta.env.VITE_APP_TITLE ?? "Jiaxiang Medical Assistant",
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? defaultApiBaseUrl,
};
