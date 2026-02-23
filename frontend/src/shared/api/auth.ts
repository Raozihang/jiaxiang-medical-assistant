import { http, setStoredToken } from "@/shared/api/http";
import type { ApiResponse } from "@/shared/types/api";

type LoginResponse = {
  token: string;
  expires_in: number;
  user: {
    name: string;
    role: string;
  };
};

export async function login(account = "doctor", password = "dev") {
  const response = await http.post<ApiResponse<LoginResponse>>("/auth/login", {
    account,
    password,
  });
  setStoredToken(response.data.data.token);
  return response.data.data;
}
