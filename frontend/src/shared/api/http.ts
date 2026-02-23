import axios from "axios";
import { env } from "@/shared/config/env";

const TOKEN_KEY = "jx_medical_token";

export function getStoredToken() {
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token);
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
  (error) => Promise.reject(error),
);
