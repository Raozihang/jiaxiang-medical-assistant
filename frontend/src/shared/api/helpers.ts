import type { Paginated } from "@/shared/types/api";

type UnknownRecord = Record<string, unknown>;

function normalizeKey(key: string) {
  return key.replace(/[_-]/g, "").toLowerCase();
}

export function asRecord(value: unknown): UnknownRecord | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  return value as UnknownRecord;
}

export function pickFirst(source: UnknownRecord | null, keys: string[]) {
  if (!source) {
    return undefined;
  }

  for (const key of keys) {
    if (key in source) {
      return source[key];
    }
  }

  const normalizedEntries = Object.entries(source).map(([key, value]) => [normalizeKey(key), value] as const);

  for (const key of keys) {
    const normalizedKey = normalizeKey(key);
    const matched = normalizedEntries.find(([candidate]) => candidate === normalizedKey);
    if (matched) {
      return matched[1];
    }
  }

  return undefined;
}

export function toNumber(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return undefined;
}

export function toBoolean(value: unknown) {
  if (typeof value === "boolean") {
    return value;
  }
  if (typeof value === "number") {
    return value !== 0;
  }
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    if (["true", "yes", "y", "1", "ok"].includes(normalized)) {
      return true;
    }
    if (["false", "no", "n", "0"].includes(normalized)) {
      return false;
    }
  }
  return undefined;
}

export function toText(value: unknown, fallback = "") {
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return fallback;
}

export function toStringArray(value: unknown) {
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === "string") {
          return item.trim();
        }
        if (typeof item === "number" || typeof item === "boolean") {
          return String(item);
        }
        const record = asRecord(item);
        if (record) {
          return (
            toText(pickFirst(record, ["name", "label", "title", "message", "text"])) ||
            JSON.stringify(item)
          );
        }
        return "";
      })
      .filter((item) => item.length > 0);
  }

  if (typeof value === "string") {
    return value
      .split(/[\n,;|]/g)
      .map((item) => item.trim())
      .filter((item) => item.length > 0);
  }

  return [];
}

export function unwrapApiData<T = unknown>(value: unknown): T {
  const record = asRecord(value);
  if (!record) {
    return value as T;
  }

  const nested = pickFirst(record, ["data", "result", "payload", "item"]);
  if (nested !== undefined) {
    return nested as T;
  }

  return value as T;
}

export function normalizePaginated<T>(
  value: unknown,
  mapItem: (item: unknown, index: number) => T,
): Paginated<T> {
  const payload = unwrapApiData<unknown>(value);
  const payloadRecord = asRecord(payload);

  const listCandidate = payloadRecord
    ? pickFirst(payloadRecord, ["items", "list", "records", "rows", "results", "data"])
    : payload;
  const list = Array.isArray(listCandidate) ? listCandidate : [];

  const page =
    toNumber(payloadRecord ? pickFirst(payloadRecord, ["page", "current", "page_index"]) : undefined) ??
    1;
  const pageSize =
    toNumber(
      payloadRecord
        ? pickFirst(payloadRecord, ["page_size", "pageSize", "size", "per_page"])
        : undefined,
    ) ?? list.length;
  const total =
    toNumber(payloadRecord ? pickFirst(payloadRecord, ["total", "count", "total_count"]) : undefined) ??
    list.length;

  return {
    items: list.map(mapItem),
    page,
    page_size: pageSize,
    total,
  };
}

export function getErrorMessage(error: unknown, fallback = "请求失败") {
  if (typeof error === "string") {
    return error;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  const errorRecord = asRecord(error);
  if (!errorRecord) {
    return fallback;
  }

  const responseRecord = asRecord(pickFirst(errorRecord, ["response"]));
  const responseData = responseRecord ? pickFirst(responseRecord, ["data"]) : undefined;
  const responseDataRecord = asRecord(responseData);
  if (responseDataRecord) {
    const apiMessage = toText(pickFirst(responseDataRecord, ["message", "error", "detail"]));
    if (apiMessage) {
      return apiMessage;
    }
  }

  const directMessage = toText(pickFirst(errorRecord, ["message", "error"]));
  return directMessage || fallback;
}

