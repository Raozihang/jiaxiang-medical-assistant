import { http } from "@/shared/api/http";
import {
  asRecord,
  normalizePaginated,
  pickFirst,
  toNumber,
  toText,
  toStringArray,
  unwrapApiData,
} from "@/shared/api/helpers";

export type VisitImportItem = {
  student_id: string;
  symptoms?: string[];
  description?: string;
  diagnosis?: string;
  prescription?: string[];
  destination?: string;
  created_at?: string;
};

export type ImportTaskError = {
  index: number;
  message: string;
};

export type ImportTask = {
  id: string;
  status: string;
  total: number;
  success: number;
  failed: number;
  progress: number;
  errors: ImportTaskError[];
  created_at: string;
  updated_at: string;
};

function toImportTaskError(item: unknown, index = 0): ImportTaskError {
  const record = asRecord(item);
  return {
    index: toNumber(pickFirst(record, ["index"])) ?? index,
    message: toText(pickFirst(record, ["message", "error"]), "未知错误"),
  };
}

function toImportTask(item: unknown): ImportTask {
  const record = asRecord(item);
  const total = toNumber(pickFirst(record, ["total"])) ?? 0;
  const success = toNumber(pickFirst(record, ["success"])) ?? 0;
  const failed = toNumber(pickFirst(record, ["failed"])) ?? 0;

  const progress =
    toNumber(pickFirst(record, ["progress"])) ??
    (total > 0 ? Math.round(((success + failed) / total) * 100) : 0);

  const errorSource = pickFirst(record, ["errors"]);
  const errors = Array.isArray(errorSource)
    ? errorSource.map(toImportTaskError)
    : toStringArray(errorSource).map((message, index) => ({ index, message }));

  return {
    id: toText(pickFirst(record, ["id"])),
    status: toText(pickFirst(record, ["status"]), "processing"),
    total,
    success,
    failed,
    progress,
    errors,
    created_at: toText(pickFirst(record, ["created_at"])),
    updated_at: toText(pickFirst(record, ["updated_at"])),
  };
}

export async function listImportTasks(params: { page?: number; pageSize?: number }) {
  const response = await http.get("/import/tasks", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 10,
    },
  });

  return normalizePaginated(response.data, (item) => toImportTask(item));
}

export async function createImportTask(items: VisitImportItem[]) {
  const response = await http.post("/import/visits", items);
  return toImportTask(unwrapApiData(response.data));
}

export async function getImportTask(id: string) {
  const response = await http.get(`/import/tasks/${id}`);
  return toImportTask(unwrapApiData(response.data));
}
