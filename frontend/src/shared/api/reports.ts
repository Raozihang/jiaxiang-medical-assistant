import { http } from "@/shared/api/http";
import {
  asRecord,
  pickFirst,
  toNumber,
  toText,
  unwrapApiData,
} from "@/shared/api/helpers";

export type OverviewReport = {
  today_visits: number;
  observation_students: number;
  stock_warnings: number;
  due_follow_ups: number;
};

export type ReportSummary = {
  totalVisits: number;
  urgentVisits: number;
  observationStudents: number;
  hospitalReferrals: number;
  returnClassCount: number;
  stockWarnings: number;
};

export type ReportTrend = {
  label: string;
  visits: number;
  urgent: number;
  observation: number;
};

export type ReportRankItem = {
  name: string;
  count: number;
};

export type PeriodReport = {
  period: string;
  generatedAt: string;
  summary: ReportSummary;
  trends: ReportTrend[];
  topSymptoms: ReportRankItem[];
  topMedicines: ReportRankItem[];
  raw: unknown;
};

const defaultSummary: ReportSummary = {
  totalVisits: 0,
  urgentVisits: 0,
  observationStudents: 0,
  hospitalReferrals: 0,
  returnClassCount: 0,
  stockWarnings: 0,
};

function toRankItem(item: unknown, index = 0): ReportRankItem {
  const record = asRecord(item);

  return {
    name: toText(pickFirst(record, ["name", "label", "symptom", "medicine"]), `项目 ${index + 1}`),
    count: toNumber(pickFirst(record, ["count", "value", "total"])) ?? 0,
  };
}

function toTrendItem(item: unknown, index = 0): ReportTrend {
  const record = asRecord(item);

  return {
    label: toText(
      pickFirst(record, ["label", "date", "day", "week", "month", "period"]),
      `时段 ${index + 1}`,
    ),
    visits: toNumber(pickFirst(record, ["visits", "visit_count", "total"])) ?? 0,
    urgent: toNumber(pickFirst(record, ["urgent", "urgent_visits", "urgent_count"])) ?? 0,
    observation:
      toNumber(
        pickFirst(record, [
          "observation",
          "observation_students",
          "observation_visits",
          "observationVisits",
          "observation_count",
        ]),
      ) ?? 0,
  };
}

function parseSummary(record: Record<string, unknown> | null): ReportSummary {
  return {
    totalVisits: toNumber(pickFirst(record, ["total_visits", "totalVisits", "visit_count"])) ?? 0,
    urgentVisits: toNumber(pickFirst(record, ["urgent_visits", "urgentVisits", "urgent_count"])) ?? 0,
    observationStudents:
      toNumber(
        pickFirst(record, [
          "observation_students",
          "observation_visits",
          "observationStudents",
          "observationVisits",
          "observation_count",
        ]),
      ) ?? 0,
    hospitalReferrals:
      toNumber(
        pickFirst(record, ["hospital_referrals", "hospitalReferrals", "hospital_count"]),
      ) ?? 0,
    returnClassCount:
      toNumber(pickFirst(record, ["return_class_count", "returnClassCount", "returned_count"])) ?? 0,
    stockWarnings: toNumber(pickFirst(record, ["stock_warnings", "stockWarnings"])) ?? 0,
  };
}

function parsePeriodReport(value: unknown, period: string): PeriodReport {
  const data = unwrapApiData<unknown>(value);
  const record = asRecord(data);
  const summaryRecord = asRecord(pickFirst(record, ["summary", "overview", "stats"])) ?? record;

  const trendSource = pickFirst(record, ["trends", "trend", "timeline", "chart"]);
  const symptomSource = pickFirst(record, ["top_symptoms", "symptoms", "symptom_ranking"]);
  const medicineSource = pickFirst(record, ["top_medicines", "medicines", "medicine_ranking"]);

  return {
    period: toText(pickFirst(record, ["period", "range"]), period),
    generatedAt:
      toText(pickFirst(record, ["generated_at", "generatedAt", "updated_at", "timestamp"])) ||
      new Date().toISOString(),
    summary: parseSummary(summaryRecord) ?? defaultSummary,
    trends: Array.isArray(trendSource) ? trendSource.map(toTrendItem) : [],
    topSymptoms: Array.isArray(symptomSource) ? symptomSource.map(toRankItem) : [],
    topMedicines: Array.isArray(medicineSource) ? medicineSource.map(toRankItem) : [],
    raw: data,
  };
}

function parseOverview(value: unknown): OverviewReport {
  const data = unwrapApiData<unknown>(value);
  const record = asRecord(data);

  return {
    today_visits:
      toNumber(pickFirst(record, ["today_visits", "todayVisits", "visit_count", "total_visits"])) ?? 0,
    observation_students:
      toNumber(
        pickFirst(record, [
          "observation_students",
          "observation_visits",
          "observationStudents",
          "observationVisits",
          "observation_count",
        ]),
      ) ?? 0,
    stock_warnings: toNumber(pickFirst(record, ["stock_warnings", "stockWarnings"])) ?? 0,
    due_follow_ups:
      toNumber(pickFirst(record, ["due_follow_ups", "dueFollowUps", "follow_up_due"])) ?? 0,
  };
}

export async function getOverviewReport() {
  const response = await http.get("/reports/overview");
  return parseOverview(response.data);
}

export async function getDailyReport(params?: { date?: string }) {
  const response = await http.get("/reports/daily", { params });
  return parsePeriodReport(response.data, "daily");
}

export async function getWeeklyReport(params?: { week?: string }) {
  const response = await http.get("/reports/weekly", { params });
  return parsePeriodReport(response.data, "weekly");
}

export async function getMonthlyReport(params?: { month?: string }) {
  const response = await http.get("/reports/monthly", { params });
  return parsePeriodReport(response.data, "monthly");
}

export async function exportReportExcel(period: "daily" | "weekly" | "monthly") {
  const response = await http.get(`/reports/export/${period}`, {
    responseType: "blob",
    timeout: 30000,
  });

  const disposition = response.headers["content-disposition"] ?? "";
  const match = disposition.match(/filename="?([^"]+)"?/);
  const filename = match?.[1] ?? `报表_${period}.xlsx`;

  const url = URL.createObjectURL(response.data as Blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = decodeURIComponent(filename);
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

