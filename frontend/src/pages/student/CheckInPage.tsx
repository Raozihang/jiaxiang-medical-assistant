import {
  AlertOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  FileTextOutlined,
  FireOutlined,
  HeartOutlined,
  IdcardOutlined,
  LoadingOutlined,
  SendOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { Alert, Button, Card, Input, message, Progress, Space, Tag, Typography } from "antd";
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  connectCheckInRealtime,
  type RealtimeMessage,
  sendCheckInProgress,
  sendRealtimeMessage,
  type TemperatureRealtimePayload,
} from "@/shared/api/realtime";
import { createVisit } from "@/shared/api/visits";

const THERMOMETER_SECONDS = 300;
const TOTAL_THERMOMETERS = 8;

type StepId = "identity" | "symptoms" | "temperature" | "description" | "confirm";
type TemperatureStatus = "normal" | "requested" | "timing" | "due" | "measured";

const stepItems: Array<{ id: StepId; title: string; description: string; icon: ReactNode }> = [
  {
    id: "identity",
    title: "登记身份",
    description: "刷学生卡或填写学号",
    icon: <IdcardOutlined />,
  },
  {
    id: "symptoms",
    title: "描述不适",
    description: "选部位或一句话描述",
    icon: <HeartOutlined />,
  },
  {
    id: "temperature",
    title: "体温判断",
    description: "必要时使用水银体温计",
    icon: <ClockCircleOutlined />,
  },
  {
    id: "description",
    title: "补充说明",
    description: "可选补充细节",
    icon: <FileTextOutlined />,
  },
  {
    id: "confirm",
    title: "确认提交",
    description: "生成医生端队列",
    icon: <CheckCircleOutlined />,
  },
];

const symptomOptions = [
  { label: "发热", value: "fever", icon: <FireOutlined /> },
  { label: "外伤", value: "injury", icon: <AlertOutlined /> },
  { label: "过敏", value: "allergy", icon: <HeartOutlined /> },
  { label: "说不清", value: "other", icon: <FileTextOutlined /> },
];

const bodyPartOptions = [
  { label: "头部", value: "head_discomfort", hint: "头痛、头晕、眼鼻耳不适" },
  { label: "颈部/咽喉", value: "neck_throat_discomfort", hint: "咽痛、颈部疼痛、吞咽不适" },
  { label: "胸部", value: "chest_discomfort", hint: "胸闷、胸痛、心慌" },
  { label: "腹部", value: "stomachache", hint: "腹痛、恶心、腹泻" },
  { label: "手臂/手部", value: "hand_arm_discomfort", hint: "手部疼痛、红肿、擦伤" },
  { label: "腿脚", value: "leg_foot_discomfort", hint: "膝踝疼痛、扭伤、行走不适" },
  { label: "皮肤", value: "skin_discomfort", hint: "皮疹、瘙痒、红肿" },
  { label: "全身", value: "body_discomfort", hint: "乏力、酸痛、说不清的不舒服" },
];

const bodyMapHitAreas = [
  { label: "头部", value: "head_discomfort", className: "head" },
  { label: "颈部/咽喉", value: "neck_throat_discomfort", className: "neck" },
  { label: "胸部", value: "chest_discomfort", className: "chest" },
  { label: "腹部", value: "stomachache", className: "stomach" },
  { label: "左手臂/手部", value: "hand_arm_discomfort", className: "left-hand" },
  { label: "右手臂/手部", value: "hand_arm_discomfort", className: "right-hand" },
  { label: "腿脚", value: "leg_foot_discomfort", className: "leg" },
  { label: "皮肤", value: "skin_discomfort", className: "skin" },
  { label: "全身", value: "body_discomfort", className: "whole" },
];

const questionPrompts = [
  "哪里最不舒服？",
  "大概持续了多久？",
  "有没有吃过药或做过处理？",
  "是否有同学陪同或老师要求来检查？",
];

function createSessionId() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function formatCountdown(seconds: number) {
  const safeSeconds = Math.max(0, seconds);
  const minutes = Math.floor(safeSeconds / 60);
  const remainder = safeSeconds % 60;
  return `${minutes}:${String(remainder).padStart(2, "0")}`;
}

function thermometerFromSession(sessionId: string) {
  const total = Array.from(sessionId).reduce((sum, char) => sum + char.charCodeAt(0), 0);
  return (total % TOTAL_THERMOMETERS) + 1;
}

function getQuestionForStep(step: StepId) {
  switch (step) {
    case "identity":
      return "请刷学生卡，或直接填写学号。";
    case "symptoms":
      return "可以选择不舒服的部位，也可以用一句话描述。";
    case "temperature":
      return "是否需要使用水银体温计测量体温？";
    case "description":
      return "请补充告诉医生，哪里不舒服、持续多久。";
    case "confirm":
      return "请确认登记内容，提交后医生端会自动收到。";
  }
}

function temperatureText(status: TemperatureStatus, value: number | null) {
  if (status === "measured" && value !== null) {
    return `${value.toFixed(1)}℃`;
  }
  if (status === "timing") {
    return "体温计测量中";
  }
  if (status === "due") {
    return "等待医生读数";
  }
  return "体温正常";
}

export function CheckInPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const socketRef = useRef<WebSocket | null>(null);
  const [sessionId, setSessionId] = useState(() => createSessionId());
  const [step, setStep] = useState<StepId>("identity");
  const [studentId, setStudentId] = useState("");
  const [selectedSymptoms, setSelectedSymptoms] = useState<string[]>([]);
  const [description, setDescription] = useState("");
  const [temperatureStatus, setTemperatureStatus] = useState<TemperatureStatus>("normal");
  const [temperatureValue, setTemperatureValue] = useState<number | null>(null);
  const [thermometerId, setThermometerId] = useState<number | null>(null);
  const [countdownSeconds, setCountdownSeconds] = useState(THERMOMETER_SECONDS);
  const [submitting, setSubmitting] = useState(false);
  const identityRef = useRef({ sessionId: "", studentId: "" });

  const studentName = studentId.trim() ? `学生-${studentId.trim()}` : "";
  const activeStepIndex = stepItems.findIndex((item) => item.id === step);
  const hasFeverComplaint = selectedSymptoms.includes("fever");
  const shouldUseThermometer =
    hasFeverComplaint ||
    temperatureStatus === "requested" ||
    temperatureStatus === "timing" ||
    temperatureStatus === "due" ||
    temperatureStatus === "measured";

  const selectedSymptomLabels = useMemo(
    () =>
      [...bodyPartOptions, ...symptomOptions]
        .filter((item) => selectedSymptoms.includes(item.value))
        .map((item) => item.label),
    [selectedSymptoms],
  );

  const intakeSummary = useMemo(() => {
    const parts = selectedSymptomLabels.length ? selectedSymptomLabels.join("、") : "";
    const text = description.trim();
    if (parts && text) {
      return `${parts}；${text}`;
    }
    return parts || text;
  }, [description, selectedSymptomLabels]);

  const emitProgress = useCallback(
    (overrides?: {
      status?: "collecting" | "temperature" | "submitted" | "cancelled";
      temperatureStatus?: TemperatureStatus;
      countdownSeconds?: number;
      temperatureValue?: number | null;
      targetStep?: StepId;
      thermometerId?: number | null;
    }) => {
      const targetStep = overrides?.targetStep ?? step;
      const targetTemperatureStatus = overrides?.temperatureStatus ?? temperatureStatus;
      sendCheckInProgress(socketRef.current, {
        session_id: sessionId,
        student_id: studentId.trim(),
        student_name: studentName,
        question_id: targetStep,
        question: getQuestionForStep(targetStep),
        symptoms: selectedSymptoms,
        description,
        temperature: overrides?.temperatureValue ?? temperatureValue,
        temperature_status: targetTemperatureStatus,
        thermometer_id: overrides?.thermometerId ?? thermometerId,
        countdown_seconds: overrides?.countdownSeconds ?? countdownSeconds,
        status:
          overrides?.status ??
          (targetStep === "temperature" || targetTemperatureStatus === "timing"
            ? "temperature"
            : "collecting"),
      });
    },
    [
      countdownSeconds,
      description,
      selectedSymptoms,
      sessionId,
      step,
      studentId,
      studentName,
      temperatureStatus,
      temperatureValue,
      thermometerId,
    ],
  );
  const emitProgressRef = useRef(emitProgress);

  useEffect(() => {
    identityRef.current = { sessionId, studentId: studentId.trim() };
  }, [sessionId, studentId]);

  useEffect(() => {
    emitProgressRef.current = emitProgress;
  }, [emitProgress]);

  const startThermometer = useCallback(
    (nextThermometerId?: number | null) => {
      const assignedThermometerId = nextThermometerId ?? thermometerFromSession(sessionId);
      setThermometerId(assignedThermometerId);
      setTemperatureStatus("timing");
      setTemperatureValue(null);
      setCountdownSeconds(THERMOMETER_SECONDS);
      emitProgress({
        targetStep: "temperature",
        status: "temperature",
        temperatureStatus: "timing",
        countdownSeconds: THERMOMETER_SECONDS,
        temperatureValue: null,
        thermometerId: assignedThermometerId,
      });
    },
    [emitProgress, sessionId],
  );

  const handleRealtimeMessage = useCallback(
    (event: MessageEvent) => {
      let message: RealtimeMessage<TemperatureRealtimePayload>;
      try {
        message = JSON.parse(event.data) as RealtimeMessage<TemperatureRealtimePayload>;
      } catch {
        return;
      }

      if (message.type !== "temperature_requested" && message.type !== "temperature_recorded") {
        return;
      }

      const payload = message.payload;
      if (!payload) {
        return;
      }

      const currentIdentity = identityRef.current;
      const matchesCurrentStudent =
        payload.session_id === currentIdentity.sessionId ||
        (payload.student_id.trim() !== "" && payload.student_id === currentIdentity.studentId);
      if (!matchesCurrentStudent) {
        return;
      }

      if (message.type === "temperature_requested") {
        const assignedThermometerId =
          payload.thermometer_id ?? thermometerFromSession(currentIdentity.sessionId);
        setStep("temperature");
        setThermometerId(assignedThermometerId);
        setTemperatureStatus("timing");
        setTemperatureValue(null);
        setCountdownSeconds(THERMOMETER_SECONDS);
        emitProgressRef.current({
          targetStep: "temperature",
          status: "temperature",
          temperatureStatus: "timing",
          countdownSeconds: THERMOMETER_SECONDS,
          temperatureValue: null,
          thermometerId: assignedThermometerId,
        });
        messageApi.info("医生已要求使用水银体温计，系统开始 5 分钟计时。");
        return;
      }

      if (message.type === "temperature_recorded") {
        const measuredValue = payload.temperature ?? null;
        setTemperatureValue(measuredValue);
        setTemperatureStatus(measuredValue === null ? "normal" : "measured");
        setCountdownSeconds(0);
        if (payload.thermometer_id) {
          setThermometerId(payload.thermometer_id);
        }
        messageApi.success("医生已录入体温，可以继续提交登记。");
        setStep("description");
      }
    },
    [messageApi],
  );

  useEffect(() => {
    let closedByPage = false;
    let retryTimer: number | undefined;

    const connect = () => {
      const socket = connectCheckInRealtime();
      socketRef.current = socket;
      socket.onmessage = handleRealtimeMessage;
      socket.onclose = () => {
        if (closedByPage) {
          return;
        }
        retryTimer = window.setTimeout(connect, 2000);
      };
    };

    connect();
    return () => {
      closedByPage = true;
      if (retryTimer) {
        window.clearTimeout(retryTimer);
      }
      socketRef.current?.close();
      socketRef.current = null;
    };
  }, [handleRealtimeMessage]);

  useEffect(() => {
    if (!studentId.trim()) {
      return;
    }
    emitProgress();
  }, [emitProgress, studentId]);

  useEffect(() => {
    if (step !== "temperature") {
      return;
    }
    if (!shouldUseThermometer) {
      setTemperatureStatus("normal");
      setTemperatureValue(null);
      setThermometerId(null);
      setCountdownSeconds(0);
      emitProgress({
        targetStep: "temperature",
        temperatureStatus: "normal",
        countdownSeconds: 0,
        temperatureValue: null,
        thermometerId: null,
      });
      return;
    }
    if (temperatureStatus === "normal" || temperatureStatus === "requested") {
      startThermometer(thermometerId);
    }
  }, [
    emitProgress,
    shouldUseThermometer,
    startThermometer,
    step,
    temperatureStatus,
    thermometerId,
  ]);

  useEffect(() => {
    if (temperatureStatus !== "timing") {
      return;
    }

    const timer = window.setInterval(() => {
      setCountdownSeconds((current) => {
        const next = Math.max(0, current - 1);
        if (next === 0) {
          const payload: TemperatureRealtimePayload = {
            session_id: sessionId,
            student_id: studentId.trim(),
            student_name: studentName,
            temperature: null,
            temperature_status: "due",
            thermometer_id: thermometerId,
            countdown_seconds: 0,
          };
          setTemperatureStatus("due");
          sendRealtimeMessage(socketRef.current, "temperature_due", payload);
          emitProgressRef.current({
            targetStep: "temperature",
            status: "temperature",
            temperatureStatus: "due",
            countdownSeconds: 0,
            temperatureValue: null,
            thermometerId,
          });
          return 0;
        }

        emitProgressRef.current({
          targetStep: "temperature",
          status: "temperature",
          temperatureStatus: "timing",
          countdownSeconds: next,
          temperatureValue: null,
          thermometerId,
        });
        return next;
      });
    }, 1000);

    return () => window.clearInterval(timer);
  }, [sessionId, studentId, studentName, temperatureStatus, thermometerId]);

  const goNext = () => {
    if (step === "identity") {
      if (!studentId.trim()) {
        messageApi.warning("请先刷卡或填写学号。");
        return;
      }
      setStep("symptoms");
      return;
    }
    if (step === "symptoms") {
      if (!selectedSymptoms.length && !description.trim()) {
        messageApi.warning("请选择一个不舒服的部位，或用一句话描述。");
        return;
      }
      setStep("temperature");
      return;
    }
    if (step === "temperature") {
      if (temperatureStatus === "timing" || temperatureStatus === "due") {
        messageApi.warning("请等待医生录入体温后再继续。");
        return;
      }
      setStep("description");
      return;
    }
    if (step === "description") {
      setStep("confirm");
    }
  };

  const goBack = () => {
    const currentIndex = stepItems.findIndex((item) => item.id === step);
    if (currentIndex > 0) {
      setStep(stepItems[currentIndex - 1].id);
    }
  };

  const toggleSymptom = (value: string) => {
    setSelectedSymptoms((current) => {
      if (current.includes(value)) {
        return current.filter((item) => item !== value);
      }
      return [...current, value];
    });
  };

  const appendPrompt = (prompt: string) => {
    setDescription((current) => {
      const prefix = current.trim() ? `${current.trim()}\n` : "";
      return `${prefix}${prompt} `;
    });
  };

  const resetFlow = () => {
    setSessionId(createSessionId());
    setStep("identity");
    setStudentId("");
    setSelectedSymptoms([]);
    setDescription("");
    setTemperatureStatus("normal");
    setTemperatureValue(null);
    setThermometerId(null);
    setCountdownSeconds(THERMOMETER_SECONDS);
  };

  const submitVisit = async () => {
    if (!studentId.trim()) {
      messageApi.warning("请先填写学号。");
      return;
    }
    if (temperatureStatus === "timing" || temperatureStatus === "due") {
      messageApi.warning("体温还未由医生确认，请稍等。");
      return;
    }

    setSubmitting(true);
    try {
      await createVisit({
        student_id: studentId.trim(),
        symptoms: selectedSymptoms.length ? selectedSymptoms : ["free_text"],
        description: description.trim(),
        temperature_status: temperatureStatus,
        temperature_value: temperatureValue,
      });
      emitProgress({
        targetStep: "confirm",
        status: "submitted",
        temperatureStatus,
        temperatureValue,
        countdownSeconds,
        thermometerId,
      });
      messageApi.success("登记成功，医生端已收到。");
      resetFlow();
    } catch {
      messageApi.error("登记提交失败，请联系医生处理。");
    } finally {
      setSubmitting(false);
    }
  };

  const renderIdentityStage = () => (
    <div className="guided-checkin__student-stage">
      <div className="rfid-stage" aria-hidden>
        <div className="rfid-reader-device">
          <div className="rfid-reader-device__slot" />
          <div className="rfid-reader-device__light" />
          <span>RFID 读卡器</span>
        </div>
        <div className="rfid-card-visual">
          <div className="rfid-card-visual__chip" />
          <strong>学生卡</strong>
          <small>JX SCHOOL</small>
        </div>
        <div className="rfid-contact-signal" />
        <div className="rfid-stage__status">
          <IdcardOutlined />
          <span>请将学生卡贴近读卡器，或在右侧填写学号</span>
        </div>
      </div>
      <div className="guided-checkin__student-form">
        <Typography.Title level={4}>请先确认是哪位学生来登记？</Typography.Title>
        <Typography.Paragraph className="guided-checkin__copy">
          可以刷 RFID 学生卡，也可以手动输入学号。确认后，系统会进入图形化问诊引导。
        </Typography.Paragraph>
        <Input
          size="large"
          prefix={<UserOutlined />}
          placeholder="请输入学号"
          value={studentId}
          onChange={(event) => setStudentId(event.target.value)}
          onPressEnter={goNext}
        />
        <Button type="primary" size="large" onClick={goNext}>
          确认并开始问诊
        </Button>
      </div>
    </div>
  );

  const renderSymptomsStage = () => (
    <div className="guided-checkin__graph-stage">
      <div className="guided-checkin__graph-banner">
        <HeartOutlined />
        <div>
          <strong>先选身体部位，或者直接说一句话。</strong>
          <span>选择后医生端会直接收到主诉摘要；如果只写一句话，也可以继续下一步。</span>
        </div>
      </div>
      <div className="body-intake">
        <div className="body-map-card">
          <div className="body-map-stage">
            <svg
              className="body-map"
              role="img"
              aria-label="可选择头部、颈部、胸部、腹部、手臂手部、腿脚、皮肤和全身不适"
              viewBox="0 0 260 520"
            >
              <title>身体部位选择图</title>
              <g
                className={`body-map__part body-map__head ${
                  selectedSymptoms.includes("head_discomfort") ? "is-active" : ""
                }`}
              >
                <circle cx="130" cy="58" r="38" />
                <text x="130" y="62">
                  头部
                </text>
              </g>
              <g
                className={`body-map__part ${
                  selectedSymptoms.includes("neck_throat_discomfort") ? "is-active" : ""
                }`}
              >
                <rect x="112" y="94" width="36" height="36" rx="12" />
                <text x="130" y="118">
                  颈部
                </text>
              </g>
              <g
                className={`body-map__part ${
                  selectedSymptoms.includes("chest_discomfort") ? "is-active" : ""
                }`}
              >
                <path d="M82 136 C94 116 166 116 178 136 L168 250 C158 268 102 268 92 250 Z" />
                <text x="130" y="188">
                  胸部
                </text>
              </g>
              <g
                className={`body-map__part ${
                  selectedSymptoms.includes("stomachache") ? "is-active" : ""
                }`}
              >
                <path d="M94 252 C104 270 156 270 166 252 L160 326 C148 348 112 348 100 326 Z" />
                <text x="130" y="300">
                  腹部
                </text>
              </g>
              <g
                className={`body-map__part ${
                  selectedSymptoms.includes("hand_arm_discomfort") ? "is-active" : ""
                }`}
              >
                <path d="M76 148 C54 180 42 226 34 288 C32 306 48 312 56 296 C68 238 80 206 92 174 Z" />
                <path d="M184 148 C206 180 218 226 226 288 C228 306 212 312 204 296 C192 238 180 206 168 174 Z" />
                <text x="54" y="260">
                  手
                </text>
                <text x="206" y="260">
                  手
                </text>
              </g>
              <g
                className={`body-map__part ${
                  selectedSymptoms.includes("leg_foot_discomfort") ? "is-active" : ""
                }`}
              >
                <path d="M102 330 C116 336 126 336 130 330 L126 470 C124 490 92 490 94 468 Z" />
                <path d="M130 330 C134 336 144 336 158 330 L166 468 C168 490 136 490 134 470 Z" />
                <text x="130" y="420">
                  腿脚
                </text>
              </g>
              <g
                className={`body-map__part body-map__skin ${
                  selectedSymptoms.includes("skin_discomfort") ? "is-active" : ""
                }`}
              >
                <circle cx="206" cy="116" r="14" />
                <circle cx="222" cy="140" r="10" />
                <circle cx="198" cy="154" r="8" />
                <text x="212" y="190">
                  皮肤
                </text>
              </g>
              <g
                className={`body-map__part body-map__whole ${
                  selectedSymptoms.includes("body_discomfort") ? "is-active" : ""
                }`}
              >
                <rect x="28" y="24" width="204" height="470" rx="42" />
                <text x="130" y="506">
                  全身
                </text>
              </g>
            </svg>
            {bodyMapHitAreas.map((item) => (
              <button
                aria-label={`选择${item.label}`}
                className={`body-map-hit body-map-hit--${item.className} ${
                  selectedSymptoms.includes(item.value) ? "is-active" : ""
                }`}
                key={`${item.value}-${item.className}`}
                type="button"
                onClick={() => toggleSymptom(item.value)}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>
        <div className="body-intake__side">
          <div className="one-line-complaint">
            <strong>一句话描述</strong>
            <Input.TextArea
              rows={4}
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              placeholder="例如：从上午开始咽痛、头晕，喝水后没有明显缓解。"
            />
          </div>
          <div className="body-part-list">
            {bodyPartOptions.map((item) => (
              <button
                className={`body-part-list__item ${
                  selectedSymptoms.includes(item.value) ? "is-active" : ""
                }`}
                key={item.value}
                type="button"
                onClick={() => toggleSymptom(item.value)}
              >
                <strong>{item.label}</strong>
                <span>{item.hint}</span>
              </button>
            ))}
          </div>
          <div className="symptom-grid symptom-grid--compact">
            {symptomOptions.map((item) => (
              <button
                className={`symptom-grid__item ${
                  selectedSymptoms.includes(item.value) ? "is-active" : ""
                }`}
                key={item.value}
                type="button"
                onClick={() => toggleSymptom(item.value)}
              >
                <span className="symptom-grid__icon">{item.icon}</span>
                <span className="symptom-grid__label">{item.label}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
      <div className="guided-checkin__choice-preview">
        <strong>医生端将看到</strong>
        <div className="guided-checkin__choice-tags">
          {intakeSummary ? (
            selectedSymptomLabels.map((label) => <Tag key={label}>{label}</Tag>)
          ) : (
            <Tag color="gold">等待选择或填写描述</Tag>
          )}
          {description.trim() ? <Tag color="blue">一句话描述已填写</Tag> : null}
        </div>
        {description.trim() ? (
          <Typography.Paragraph className="guided-checkin__choice-text">
            {description.trim()}
          </Typography.Paragraph>
        ) : null}
      </div>
    </div>
  );

  const renderTemperatureStage = () => {
    const percent = shouldUseThermometer
      ? Math.round(((THERMOMETER_SECONDS - countdownSeconds) / THERMOMETER_SECONDS) * 100)
      : 100;

    return (
      <div className="guided-checkin__temperature-stage">
        <div className="temperature-stage">
          <div className="temperature-stage__gauge" aria-hidden>
            <div className="temperature-stage__track" />
            <div
              className="temperature-stage__fill"
              style={{ height: `${Math.max(6, percent)}%` }}
            />
            <span className="temperature-stage__tick temperature-stage__tick--top" />
            <span className="temperature-stage__tick temperature-stage__tick--mid" />
            <span className="temperature-stage__tick temperature-stage__tick--bottom" />
          </div>
          <div className="temperature-stage__meta">
            <Typography.Title level={4}>水银体温计测量</Typography.Title>
            {shouldUseThermometer ? (
              <>
                <Space wrap>
                  <Tag color="orange">体温计 {thermometerId ?? "-"}</Tag>
                  <Tag color={temperatureStatus === "measured" ? "green" : "gold"}>
                    {temperatureText(temperatureStatus, temperatureValue)}
                  </Tag>
                </Space>
                <div className="temperature-stage__countdown">
                  {temperatureStatus === "timing" ? <LoadingOutlined /> : <ClockCircleOutlined />}
                  <strong>{formatCountdown(countdownSeconds)}</strong>
                </div>
                <Progress percent={percent} showInfo={false} status="active" />
                {temperatureStatus === "due" ? (
                  <Alert
                    type="warning"
                    showIcon
                    message="5 分钟已到"
                    description="请把体温计交给医生读数，医生录入后会自动进入下一步。"
                  />
                ) : null}
                {temperatureStatus === "measured" && temperatureValue !== null ? (
                  <Alert
                    type="success"
                    showIcon
                    message={`医生已录入体温：${temperatureValue.toFixed(1)}℃`}
                  />
                ) : null}
              </>
            ) : (
              <Alert
                type="success"
                showIcon
                message="体温正常"
                description="学生没有主诉发热，也没有医生要求测温，本次登记将记录为体温正常。"
              />
            )}
          </div>
        </div>
      </div>
    );
  };

  const renderDescriptionStage = () => (
    <div className="guided-checkin__note-stage">
      <div className="note-stage">
        <div>
          <strong>如果还有细节，可以继续补充。</strong>
          <span>前一步已经能完成主诉采集，这里只记录持续时间、用药等额外信息。</span>
        </div>
        <Space wrap>
          {questionPrompts.map((prompt) => (
            <Button key={prompt} onClick={() => appendPrompt(prompt)}>
              {prompt}
            </Button>
          ))}
        </Space>
        <Input.TextArea
          rows={6}
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          placeholder="例如：上午体育课后开始头晕，喝水后没有明显缓解。"
        />
      </div>
    </div>
  );

  const renderConfirmStage = () => (
    <div className="guided-checkin__summary-stage">
      <div className="summary-stage">
        <CheckCircleOutlined />
        <div>
          <strong>请确认登记内容</strong>
          <span>提交后医生端会自动收到学生队列和体温状态。</span>
        </div>
      </div>
      <div className="guided-checkin__summary">
        <Tag color="blue">学号：{studentId.trim()}</Tag>
        <Tag color={hasFeverComplaint ? "red" : "green"}>主诉：{intakeSummary || "无明显主诉"}</Tag>
        <Tag color={temperatureStatus === "measured" ? "orange" : "green"}>
          体温：{temperatureText(temperatureStatus, temperatureValue)}
        </Tag>
        <Typography.Paragraph>
          {description.trim() || "未填写一句话描述，医生端会按已选部位处理。"}
        </Typography.Paragraph>
      </div>
    </div>
  );

  const renderStage = () => {
    switch (step) {
      case "identity":
        return renderIdentityStage();
      case "symptoms":
        return renderSymptomsStage();
      case "temperature":
        return renderTemperatureStage();
      case "description":
        return renderDescriptionStage();
      case "confirm":
        return renderConfirmStage();
    }
  };

  return (
    <Space className="guided-checkin" direction="vertical" size={18}>
      {contextHolder}
      <div className="guided-checkin__agent">
        <span className="guided-checkin__eyebrow">自助登记引导</span>
        <div className="guided-checkin__header">
          <Typography.Title level={3} style={{ margin: 0 }}>
            {getQuestionForStep(step)}
          </Typography.Title>
          <p className="guided-checkin__copy">
            系统会把学生登记进度、体温计倒计时和提交结果实时同步到医生端。
          </p>
        </div>
      </div>

      <div className="guided-checkin__rail">
        {stepItems.map((item, index) => (
          <div
            className={`guided-checkin__rail-item ${item.id === step ? "is-active" : ""} ${
              index < activeStepIndex ? "is-done" : ""
            }`}
            key={item.id}
          >
            <span className="guided-checkin__rail-icon">{item.icon}</span>
            <strong>{item.title}</strong>
            <small>{item.description}</small>
          </div>
        ))}
      </div>

      <Card className="guided-checkin__panel">{renderStage()}</Card>

      <div className="guided-checkin__actions">
        <Button disabled={step === "identity"} onClick={goBack}>
          上一步
        </Button>
        {step === "confirm" ? (
          <Button
            icon={<SendOutlined />}
            loading={submitting}
            type="primary"
            onClick={() => void submitVisit()}
          >
            提交登记
          </Button>
        ) : (
          <Button type="primary" onClick={goNext}>
            下一步
          </Button>
        )}
      </div>
    </Space>
  );
}
