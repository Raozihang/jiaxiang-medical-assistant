import { http } from "@/shared/api/http";
import {
  asRecord,
  pickFirst,
  toBoolean,
  toNumber,
  toText,
  toStringArray,
  unwrapApiData,
} from "@/shared/api/helpers";

export type AnalyzeSymptomsPayload = {
  visit_id?: string;
  symptoms: string[];
  description?: string;
};

export type StructuredSymptom = {
  name: string;
  severity?: string;
  duration?: string;
  note?: string;
  confidence?: number;
};

export type AnalyzeResult = {
  summary: string;
  structuredSymptoms: StructuredSymptom[];
  possibleConditions: string[];
  riskFlags: string[];
  recommendations: string[];
  raw: unknown;
};

export type TriagePayload = {
  visit_id?: string;
  symptoms: string[];
  description?: string;
  analysis_summary?: string;
};

export type TriageResult = {
  level: string;
  destination: string;
  reason: string;
  recommendations: string[];
  riskFlags: string[];
  raw: unknown;
};

export type RecommendPayload = {
  visit_id?: string;
  symptoms: string[];
  diagnosis?: string;
  triage_level?: string;
  destination?: string;
  allergies?: string[];
};

export type MedicineRecommendation = {
  name: string;
  dosage: string;
  frequency: string;
  duration: string;
  reason: string;
  caution: string;
};

export type RecommendResult = {
  medicines: MedicineRecommendation[];
  advice: string[];
  contraindications: string[];
  raw: unknown;
};

export type InteractionCheckPayload = {
  medicines: string[];
  student_id?: string;
};

export type InteractionWarning = {
  title: string;
  severity: string;
  description: string;
  suggestion: string;
};

export type InteractionCheckResult = {
  hasInteraction: boolean;
  severity: string;
  warnings: InteractionWarning[];
  safe: boolean;
  raw: unknown;
};

function parseStructuredSymptom(item: unknown): StructuredSymptom {
  const record = asRecord(item);
  if (!record) {
    const text = toText(item, "未知症状");
    return { name: text };
  }

  const severityNumber = toNumber(pickFirst(record, ["severity", "priority", "level"]));
  return {
    name: toText(pickFirst(record, ["name", "symptom", "label"]), "未知症状"),
    severity:
      toText(pickFirst(record, ["severity_text", "severity", "level"])) ||
      (severityNumber !== undefined ? String(severityNumber) : undefined),
    duration: toText(pickFirst(record, ["duration", "duration_text", "lasting"])) || undefined,
    note: toText(pickFirst(record, ["note", "description", "remark"])) || undefined,
    confidence: toNumber(pickFirst(record, ["confidence", "score"])),
  };
}

function parseMedicineRecommendation(item: unknown): MedicineRecommendation {
  const record = asRecord(item);
  if (!record) {
    return {
      name: toText(item, "未命名药品"),
      dosage: "",
      frequency: "",
      duration: "",
      reason: "",
      caution: "",
    };
  }

  return {
    name: toText(pickFirst(record, ["name", "medicine", "medicine_name", "drug"]), "未命名药品"),
    dosage: toText(pickFirst(record, ["dosage", "dose"])),
    frequency: toText(pickFirst(record, ["frequency", "usage", "times"])),
    duration: toText(pickFirst(record, ["duration", "course"])),
    reason: toText(pickFirst(record, ["reason", "indication", "why"])),
    caution: toText(pickFirst(record, ["caution", "warning", "note"])),
  };
}

function parseInteractionWarning(item: unknown): InteractionWarning {
  const record = asRecord(item);
  if (!record) {
    const text = toText(item, "可能存在药物相互作用");
    return {
      title: text,
      severity: "medium",
      description: text,
      suggestion: "请结合临床判断复核用药方案",
    };
  }

  return {
    title:
      toText(pickFirst(record, ["title", "name", "interaction"])) ||
      toStringArray(pickFirst(record, ["pair"])).join(" + ") ||
      "药物相互作用",
    severity: toText(pickFirst(record, ["severity", "level", "risk"]), "medium"),
    description:
      toText(pickFirst(record, ["description", "message", "detail", "effect"])) ||
      "请关注药物联用风险",
    suggestion:
      toText(pickFirst(record, ["suggestion", "advice", "recommendation"])) ||
      "建议调整药物组合或加强观察",
  };
}

function parseAnalyzeResult(payload: unknown): AnalyzeResult {
  const data = unwrapApiData<unknown>(payload);
  const record = asRecord(data);

  const structuredSource = record
    ? pickFirst(record, ["structured_symptoms", "structuredSymptoms", "symptoms", "symptom_list"])
    : [];
  const structuredSymptoms = Array.isArray(structuredSource)
    ? structuredSource.map(parseStructuredSymptom)
    : [];

  const matchedSignals = toStringArray(pickFirst(record, ["matched_signals"]));
  const riskLevel = toText(pickFirst(record, ["risk_level"]));

  return {
    summary:
      toText(pickFirst(record, ["summary", "analysis", "chief_complaint"])) ||
      (riskLevel ? `风险等级：${riskLevel}` : "暂无分析总结"),
    structuredSymptoms:
      structuredSymptoms.length > 0
        ? structuredSymptoms
        : matchedSignals.map((item) => ({ name: item, severity: riskLevel || undefined })),
    possibleConditions: toStringArray(
      pickFirst(record, ["possible_conditions", "conditions", "diagnosis_candidates", "possible_causes"]),
    ),
    riskFlags: toStringArray(pickFirst(record, ["risk_flags", "alerts", "red_flags", "matched_signals"])),
    recommendations: toStringArray(
      pickFirst(record, ["recommendations", "advice", "next_steps", "suggestions", "suggested_actions"]),
    ),
    raw: data,
  };
}

function parseTriageResult(payload: unknown): TriageResult {
  const data = unwrapApiData<unknown>(payload);
  const record = asRecord(data);

  return {
    level: toText(pickFirst(record, ["triage_level", "level", "priority"]), "normal"),
    destination: toText(
      pickFirst(record, ["destination", "suggested_destination", "recommend_destination"]),
      "observation",
    ),
    reason:
      toText(pickFirst(record, ["reason", "rationale", "summary"])) ||
      "AI 建议优先结合现场问诊确认",
    recommendations: toStringArray(
      pickFirst(record, ["recommendations", "suggestions", "actions", "next_steps", "suggested_actions"]),
    ),
    riskFlags: toStringArray(pickFirst(record, ["risk_flags", "alerts", "red_flags"])),
    raw: data,
  };
}

function parseRecommendResult(payload: unknown): RecommendResult {
  const data = unwrapApiData<unknown>(payload);
  const record = asRecord(data);

  const medicineSource = record
    ? pickFirst(record, ["medicines", "recommendations", "items", "list"])
    : [];
  const medicines = Array.isArray(medicineSource) ? medicineSource.map(parseMedicineRecommendation) : [];

  const hints = toStringArray(pickFirst(record, ["medicine_hints"]));

  return {
    medicines:
      medicines.length > 0
        ? medicines
        : hints.map((item) => ({
            name: item,
            dosage: "",
            frequency: "",
            duration: "",
            reason: "AI 推荐提示",
            caution: "",
          })),
    advice: toStringArray(pickFirst(record, ["advice", "instructions", "recommendations", "care_plan"])),
    contraindications: toStringArray(
      pickFirst(record, ["contraindications", "forbidden", "warnings"]),
    ),
    raw: data,
  };
}

function parseInteractionResult(payload: unknown): InteractionCheckResult {
  const data = unwrapApiData<unknown>(payload);
  const record = asRecord(data);

  const warningSource = record
    ? pickFirst(record, ["warnings", "interactions", "alerts", "risk_list"])
    : [];
  const warnings = Array.isArray(warningSource) ? warningSource.map(parseInteractionWarning) : [];
  const hasInteraction =
    toBoolean(pickFirst(record, ["has_interaction", "hasInteraction", "risk", "conflict"])) ??
    warnings.length > 0;

  return {
    hasInteraction,
    severity: toText(pickFirst(record, ["severity", "level", "risk_level"]), "none"),
    warnings,
    safe:
      toBoolean(pickFirst(record, ["safe", "is_safe", "compatible"])) ??
      (!hasInteraction || warnings.length === 0),
    raw: data,
  };
}

export async function analyzeSymptoms(payload: AnalyzeSymptomsPayload) {
  const response = await http.post("/ai/analyze", payload);
  return parseAnalyzeResult(response.data);
}

export async function triageVisit(payload: TriagePayload) {
  const response = await http.post("/ai/triage", payload);
  return parseTriageResult(response.data);
}

export async function recommendMedicines(payload: RecommendPayload) {
  const response = await http.post("/ai/recommend", payload);
  return parseRecommendResult(response.data);
}

export async function checkMedicineInteractions(payload: InteractionCheckPayload) {
  const response = await http.post("/ai/interaction-check", payload);
  return parseInteractionResult(response.data);
}

