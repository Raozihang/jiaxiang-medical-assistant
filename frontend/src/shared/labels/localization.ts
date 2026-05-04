function looksLikeCode(value: string) {
  return /^[A-Za-z][A-Za-z0-9_\-\s]*$/.test(value.trim());
}

export function getDestinationLabel(value: string | null | undefined) {
  const trimmed = (value ?? "").trim();
  switch (trimmed.toLowerCase()) {
    case "":
      return "未登记";
    case "observation":
      return "留观";
    case "return_class":
    case "back_to_class":
    case "classroom":
      return "返回班级";
    case "urgent":
      return "紧急处理";
    case "hospital":
      return "转诊";
    case "referred":
      return "转外院";
    case "leave_school":
      return "离校就医";
    case "back_to_dorm":
    case "dormitory":
      return "返回宿舍";
    case "home":
      return "离校回家";
    case "unknown":
      return "未登记";
    default:
      return looksLikeCode(trimmed) ? "其他去向" : trimmed;
  }
}

export function getSymptomLabel(value: string | null | undefined) {
  const trimmed = (value ?? "").trim();
  const labels: Record<string, string> = {
    fever: "发热",
    cough: "咳嗽",
    headache: "头痛",
    stomachache: "腹痛",
    dizziness: "头晕",
    injury: "外伤",
    allergy: "过敏",
    other: "其他不适",
    head_discomfort: "头部不适",
    neck_throat_discomfort: "颈部/咽喉不适",
    chest_discomfort: "胸部不适",
    hand_arm_discomfort: "手臂/手部不适",
    leg_foot_discomfort: "腿脚不适",
    skin_discomfort: "皮肤不适",
    body_discomfort: "全身不适",
    free_text: "一句话描述",
  };
  return labels[trimmed] ?? (looksLikeCode(trimmed) ? "其他症状" : trimmed);
}

export function normalizeDestinationForForm(value: string | null | undefined) {
  const normalized = (value ?? "").trim().toLowerCase();
  if (["urgent", "critical", "high"].includes(normalized)) {
    return "urgent";
  }
  if (["hospital", "referral", "transfer", "referred", "leave_school"].includes(normalized)) {
    return "hospital";
  }
  if (
    ["return_class", "returnclass", "back_to_class", "class", "classroom", "back"].includes(
      normalized,
    )
  ) {
    return "return_class";
  }
  return "observation";
}

export function getPeriodLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "daily":
      return "日报";
    case "weekly":
      return "周报";
    case "monthly":
      return "月报";
    default:
      return "报表";
  }
}

export function getStatusLabel(value: string | null | undefined) {
  const trimmed = (value ?? "").trim();
  const labels: Record<string, string> = {
    sent: "已发送",
    success: "成功",
    ok: "正常",
    connected: "已接通",
    completed: "已完成",
    completed_with_errors: "部分成功",
    failed: "失败",
    error: "错误",
    busy: "占线",
    no_answer: "未接听",
    cancelled: "已取消",
    pending: "待处理",
    queued: "排队中",
    running: "运行中",
    requested: "已请求",
    processing: "处理中",
    measured: "已测量",
    due: "待测量",
    timing: "测温中",
    normal: "正常",
    open: "未处理",
    new: "新告警",
    resolved: "已处理",
    closed: "已关闭",
    done: "已完成",
  };
  return labels[trimmed.toLowerCase()] ?? (looksLikeCode(trimmed) ? "其他状态" : trimmed);
}

export function getChannelLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "wechat":
      return "微信";
    case "dingtalk":
      return "钉钉";
    default:
      return "通知渠道";
  }
}

export function getScenarioLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "visit_completed":
      return "就诊完成";
    case "observation_notice":
      return "留观通知";
    case "follow_up_reminder":
      return "复诊提醒";
    case "external_medical_followup":
      return "外出就医跟进";
    default:
      return "通知场景";
  }
}

export function getTriggerSourceLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "system":
      return "系统触发";
    case "manual":
      return "手动触发";
    default:
      return "触发来源";
  }
}

export function getProviderLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "mock":
      return "模拟外呼";
    case "aliyun":
    case "bailian":
      return "阿里云外呼";
    default:
      return "外呼服务";
  }
}

export function getSafetyLevelLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "critical":
      return "严重";
    case "high":
      return "高";
    case "urgent":
      return "紧急";
    case "medium":
      return "中";
    case "warning":
      return "警告";
    case "low":
      return "低";
    default:
      return "一般";
  }
}

export function getSafetyTypeLabel(value: string | null | undefined) {
  switch ((value ?? "").trim().toLowerCase()) {
    case "observation_timeout":
      return "留观超时";
    case "visit_unclosed":
      return "就诊未关闭";
    case "repeat_visit_3d":
      return "3天内重复就诊";
    default:
      return "安全告警";
  }
}
