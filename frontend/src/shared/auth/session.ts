export type UserRole = "student" | "doctor" | "admin";

export type AuthUser = {
  name: string;
  role: UserRole;
};

const TOKEN_KEY = "jx_medical_token";
const USER_KEY = "jx_medical_user";

function isUserRole(value: unknown): value is UserRole {
  return value === "student" || value === "doctor" || value === "admin";
}

export function getStoredToken() {
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function getStoredUser() {
  const raw = window.localStorage.getItem(USER_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<AuthUser>;
    if (!parsed || typeof parsed.name !== "string" || !isUserRole(parsed.role)) {
      return null;
    }
    return {
      name: parsed.name,
      role: parsed.role,
    };
  } catch {
    return null;
  }
}

export function setStoredUser(user: AuthUser) {
  window.localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearAuth() {
  window.localStorage.removeItem(TOKEN_KEY);
  window.localStorage.removeItem(USER_KEY);
}

export function resolveHomePath(role: UserRole | null | undefined) {
  if (role === "admin") {
    return "/admin/dashboard";
  }
  if (role === "doctor") {
    return "/doctor/visits";
  }
  return "/student/checkin";
}

export function hasValidSession() {
  return Boolean(getStoredToken() && getStoredUser());
}
