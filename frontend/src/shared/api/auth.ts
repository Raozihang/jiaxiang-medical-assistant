import { http } from "@/shared/api/http";
import {
  setStoredToken,
  setStoredUser,
  type AuthUser,
} from "@/shared/auth/session";
import type { ApiResponse } from "@/shared/types/api";

type LoginResponse = {
  token: string;
  expires_in: number;
  user: AuthUser;
};

export async function login(account: string, password: string) {
  const response = await http.post<ApiResponse<LoginResponse>>("/auth/login", {
    account,
    password,
  });
  const data = response.data.data;
  setStoredToken(data.token);
  setStoredUser(data.user);
  return data;
}
