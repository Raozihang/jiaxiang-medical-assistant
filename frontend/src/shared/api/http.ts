import axios from "axios";
import { env } from "@/shared/config/env";
import { clearAuth, getStoredToken } from "@/shared/auth/session";

function redirectToLoginIfNeeded() {
  if (window.location.pathname !== "/login") {
    window.location.assign("/login");
  }
}

export const http = axios.create({
  baseURL: env.apiBaseUrl,
  timeout: 10000,
});

http.interceptors.request.use((config) => {
  const token = getStoredToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  } else {
    delete config.headers.Authorization;
  }
  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error?.response?.status === 401) {
      clearAuth();
      redirectToLoginIfNeeded();
    }
    return Promise.reject(error);
  },
);
