import {
  AlertOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExperimentOutlined,
  FireOutlined,
  IdcardOutlined,
  UserSwitchOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Badge,
  Button,
  Card,
  Empty,
  InputNumber,
  Modal,
  message,
  notification,
  Segmented,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getErrorMessage } from "@/shared/api/helpers";
import {
  type CheckInProgressPayload,
  connectDoctorRealtime,
  type RealtimeMessage,
  sendRealtimeMessage,
  type TemperatureRealtimePayload,
  type VisitsSnapshotPayload,
} from "@/shared/api/realtime";
import { listVisits, updateVisit, type Visit } from "@/shared/api/visits";
import {
  getDestinationLabel,
  getStatusLabel,
  getSymptomLabel,
  normalizeDestinationForForm,
} from "@/shared/labels/localization";

const TOTAL_THERMOMETERS = 8;
const THERMOMETER_SECONDS = 300;

type LiveStudent = CheckInProgressPayload & {
  updated_at: string;
};

type VisitRow = {
  id: string;
  studentName: string;
  className: string;
  symptoms: string;
  temperature: string;
  temperatureStatus: string;
  destination: string;
  isObservation: boolean;
  createdAt: string;
};

function formatCountdown(seconds?: number) {
  const safeSeconds = Math.max(0, seconds ?? 0);
  const minutes = Math.floor(safeSeconds / 60);
  const remainder = safeSeconds % 60;
  return `${minutes}:${String(remainder).padStart(2, "0")}`;
}

function toRow(visit: Visit): VisitRow {
  const destination = normalizeDestination(visit.destination);
  return {
    id: visit.id,
    studentName: visit.student_name,
    className: visit.class_name,
    symptoms: visit.symptoms.map(symptomLabel).join("、") || "-",
    temperature:
      visit.temperature_status === "measured" && visit.temperature_value !== null
        ? `${visit.temperature_value.toFixed(1)}℃`
        : "体温正常",
    temperatureStatus: visit.temperature_status,
    destination: visit.destination,
    isObservation: destination === "observation",
    createdAt: new Date(visit.created_at).toLocaleString(),
  };
}

function symptomLabel(value: string) {
  return getSymptomLabel(value);
}

function normalizeDestination(value: string | null | undefined) {
  return normalizeDestinationForForm(value);
}

function destinationLabel(value: string | null | undefined) {
  return getDestinationLabel(value);
}

function destinationColor(value: string | null | undefined) {
  switch (normalizeDestination(value)) {
    case "urgent":
      return "red";
    case "hospital":
      return "orange";
    case "return_class":
      return "green";
    default:
      return "blue";
  }
}

function temperatureColor(status: string) {
  if (status === "measured") {
    return "orange";
  }
  if (status === "due" || status === "timing" || status === "requested") {
    return "gold";
  }
  return "green";
}

function isActiveThermometer(status?: string) {
  return status === "requested" || status === "timing" || status === "due";
}

function formatObservationDuration(createdAt: string) {
  const startedAt = new Date(createdAt).getTime();
  if (!Number.isFinite(startedAt)) {
    return "时间待确认";
  }
  const minutes = Math.max(0, Math.floor((Date.now() - startedAt) / 60000));
  if (minutes < 60) {
    return `已留观 ${minutes} 分钟`;
  }
  const hours = Math.floor(minutes / 60);
  const remainder = minutes % 60;
  return `已留观 ${hours} 小时 ${remainder} 分钟`;
}

function buildTemperaturePayload(
  student: LiveStudent,
  temperature: number | null,
  status: TemperatureRealtimePayload["temperature_status"],
): TemperatureRealtimePayload {
  return {
    session_id: student.session_id,
    student_id: student.student_id,
    student_name: student.student_name,
    temperature,
    temperature_status: status,
    thermometer_id: student.thermometer_id ?? null,
    countdown_seconds: status === "timing" || status === "requested" ? THERMOMETER_SECONDS : 0,
  };
}

export function VisitsPage() {
  const navigate = useNavigate();
  const socketRef = useRef<WebSocket | null>(null);
  const dueNotifiedRef = useRef<Set<string>>(new Set());
  const [messageApi, messageHolder] = message.useMessage();
  const [modal, modalHolder] = Modal.useModal();
  const [notificationApi, notificationHolder] = notification.useNotification();
  const [visits, setVisits] = useState<Visit[]>([]);
  const [loading, setLoading] = useState(false);
  const [liveStudents, setLiveStudents] = useState<Record<string, LiveStudent>>({});
  const [temperatureInputs, setTemperatureInputs] = useState<Record<string, number | null>>({});
  const [quickVisit, setQuickVisit] = useState<Visit | null>(null);
  const [quickDestination, setQuickDestination] = useState("observation");
  const [updatingDestination, setUpdatingDestination] = useState(false);

  const loadVisits = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listVisits({ page: 1, pageSize: 100 });
      setVisits(data.items);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "就诊队列加载失败"));
    } finally {
      setLoading(false);
    }
  }, [messageApi]);

  useEffect(() => {
    void loadVisits();
  }, [loadVisits]);

  const upsertLiveStudent = useCallback((payload: CheckInProgressPayload) => {
    setLiveStudents((current) => {
      const next = { ...current };
      if (payload.status === "submitted" || payload.status === "cancelled") {
        delete next[payload.session_id];
        return next;
      }
      next[payload.session_id] = {
        ...payload,
        updated_at: new Date().toISOString(),
      };
      return next;
    });
  }, []);

  const handleDueNotification = useCallback(
    (payload: TemperatureRealtimePayload) => {
      const key = payload.session_id;
      if (dueNotifiedRef.current.has(key)) {
        return;
      }
      dueNotifiedRef.current.add(key);

      const title = `体温计 ${payload.thermometer_id ?? "-"} 已到 5 分钟`;
      const studentText = `${payload.student_name || "学生"}（${payload.student_id}）`;
      notificationApi.warning({
        key,
        message: title,
        description: `${studentText} 需要医生读取水银体温计并录入体温。`,
        duration: 0,
      });
      modal.warning({
        title,
        content: `${studentText} 的水银体温计计时已结束，请在体温计管理区录入读数。`,
        okText: "去录入",
      });
    },
    [modal, notificationApi],
  );

  const handleRealtimeMessage = useCallback(
    (event: MessageEvent) => {
      let message: RealtimeMessage;
      try {
        message = JSON.parse(event.data) as RealtimeMessage;
      } catch {
        return;
      }

      if (message.type === "visits_snapshot") {
        const payload = message.payload as VisitsSnapshotPayload | undefined;
        if (payload?.items) {
          setVisits(payload.items);
        }
        return;
      }

      if (message.type === "checkin_progress") {
        const payload = message.payload as CheckInProgressPayload | undefined;
        if (payload) {
          upsertLiveStudent(payload);
        }
        return;
      }

      if (message.type === "temperature_due") {
        const payload = message.payload as TemperatureRealtimePayload | undefined;
        if (!payload) {
          return;
        }
        setLiveStudents((current) => ({
          ...current,
          [payload.session_id]: {
            ...(current[payload.session_id] ?? {
              question_id: "temperature",
              question: "水银体温计计时已结束，请医生录入读数。",
              symptoms: [],
              description: "",
              status: "temperature",
            }),
            session_id: payload.session_id,
            student_id: payload.student_id,
            student_name: payload.student_name,
            temperature: null,
            temperature_status: "due",
            thermometer_id: payload.thermometer_id ?? null,
            countdown_seconds: 0,
            updated_at: new Date().toISOString(),
            status: "temperature",
          },
        }));
        handleDueNotification(payload);
        return;
      }

      if (message.type === "temperature_recorded") {
        const payload = message.payload as TemperatureRealtimePayload | undefined;
        if (!payload) {
          return;
        }
        notificationApi.destroy(payload.session_id);
        setLiveStudents((current) => {
          const existing = current[payload.session_id];
          if (!existing) {
            return current;
          }
          return {
            ...current,
            [payload.session_id]: {
              ...existing,
              temperature: payload.temperature ?? null,
              temperature_status: payload.temperature === null ? "normal" : "measured",
              countdown_seconds: 0,
              updated_at: new Date().toISOString(),
            },
          };
        });
      }
    },
    [handleDueNotification, notificationApi, upsertLiveStudent],
  );

  useEffect(() => {
    let closedByPage = false;
    let retryTimer: number | undefined;

    const connect = () => {
      const socket = connectDoctorRealtime();
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

  const liveStudentList = useMemo(
    () => Object.values(liveStudents).sort((a, b) => b.updated_at.localeCompare(a.updated_at)),
    [liveStudents],
  );

  const activeThermometerEntries = useMemo(
    () =>
      liveStudentList
        .filter((item) => item.thermometer_id && isActiveThermometer(item.temperature_status))
        .sort((a, b) => {
          if (a.temperature_status === "due" && b.temperature_status !== "due") {
            return -1;
          }
          if (a.temperature_status !== "due" && b.temperature_status === "due") {
            return 1;
          }
          return (a.thermometer_id ?? 0) - (b.thermometer_id ?? 0);
        }),
    [liveStudentList],
  );

  const usedThermometers = useMemo(
    () => new Set(activeThermometerEntries.map((item) => item.thermometer_id).filter(Boolean)),
    [activeThermometerEntries],
  );

  const idleThermometerCount = Math.max(0, TOTAL_THERMOMETERS - usedThermometers.size);

  const observationVisits = useMemo(
    () =>
      visits
        .filter((visit) => normalizeDestination(visit.destination) === "observation")
        .sort((a, b) => b.updated_at.localeCompare(a.updated_at)),
    [visits],
  );

  const observationWarningCount = useMemo(
    () =>
      observationVisits.filter((visit) => {
        const temp = visit.temperature_value ?? 0;
        return normalizeDestination(visit.destination) === "observation" && temp >= 37.3;
      }).length,
    [observationVisits],
  );

  const requestTemperature = (student: LiveStudent) => {
    const thermometerId =
      Array.from({ length: TOTAL_THERMOMETERS }, (_, index) => index + 1).find(
        (id) => !usedThermometers.has(id),
      ) ?? 1;
    const payload: TemperatureRealtimePayload = {
      session_id: student.session_id,
      student_id: student.student_id,
      student_name: student.student_name,
      temperature: null,
      temperature_status: "requested",
      thermometer_id: thermometerId,
      countdown_seconds: THERMOMETER_SECONDS,
    };
    sendRealtimeMessage(socketRef.current, "temperature_requested", payload);
    setLiveStudents((current) => ({
      ...current,
      [student.session_id]: {
        ...student,
        temperature_status: "requested",
        thermometer_id: thermometerId,
        countdown_seconds: THERMOMETER_SECONDS,
        status: "temperature",
        updated_at: new Date().toISOString(),
      },
    }));
    messageApi.success(
      `已要求 ${student.student_name || student.student_id} 使用体温计 ${thermometerId}`,
    );
  };

  const recordTemperature = (student: LiveStudent) => {
    const value = temperatureInputs[student.session_id];
    if (typeof value !== "number" || !Number.isFinite(value)) {
      messageApi.warning("请先输入体温读数。");
      return;
    }

    const payload = buildTemperaturePayload(student, value, "measured");
    sendRealtimeMessage(socketRef.current, "temperature_recorded", payload);
    notificationApi.destroy(student.session_id);
    setLiveStudents((current) => ({
      ...current,
      [student.session_id]: {
        ...student,
        temperature: value,
        temperature_status: "measured",
        countdown_seconds: 0,
        updated_at: new Date().toISOString(),
      },
    }));
    setTemperatureInputs((current) => ({ ...current, [student.session_id]: null }));
    messageApi.success("体温已同步给学生端。");
  };

  const openQuickDestination = (visit: Visit) => {
    setQuickVisit(visit);
    setQuickDestination(normalizeDestination(visit.destination));
  };

  const handleQuickDestinationSave = async () => {
    if (!quickVisit) {
      return;
    }

    setUpdatingDestination(true);
    try {
      const updated = await updateVisit(quickVisit.id, { destination: quickDestination });
      setVisits((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      setQuickVisit(null);
      messageApi.success(
        `已将 ${quickVisit.student_name} 标记为${destinationLabel(quickDestination)}`,
      );
      await loadVisits();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "状态调整失败"));
    } finally {
      setUpdatingDestination(false);
    }
  };

  const columns: ColumnsType<VisitRow> = [
    { title: "学生姓名", dataIndex: "studentName", width: 140 },
    { title: "班级", dataIndex: "className", width: 140 },
    { title: "主诉", dataIndex: "symptoms" },
    {
      title: "体温",
      dataIndex: "temperature",
      width: 130,
      render: (value: string, row) => (
        <Tag color={temperatureColor(row.temperatureStatus)}>{getStatusLabel(value)}</Tag>
      ),
    },
    {
      title: "去向",
      dataIndex: "destination",
      width: 120,
      render: (value: string) => (
        <Tag color={destinationColor(value)}>{destinationLabel(value)}</Tag>
      ),
    },
    { title: "登记时间", dataIndex: "createdAt", width: 190 },
    {
      title: "操作",
      width: 100,
      render: (_, row) => (
        <Button type="link" onClick={() => navigate(`/doctor/visit/${row.id}`)}>
          查看
        </Button>
      ),
    },
  ];

  return (
    <Space className="visits-workbench" direction="vertical" size={16}>
      {messageHolder}
      {modalHolder}
      {notificationHolder}
      <Modal
        title="快速调整留观状态"
        open={Boolean(quickVisit)}
        confirmLoading={updatingDestination}
        okText="保存状态"
        cancelText="取消"
        onOk={() => void handleQuickDestinationSave()}
        onCancel={() => setQuickVisit(null)}
      >
        {quickVisit ? (
          <Space direction="vertical" size={14} style={{ width: "100%" }}>
            <div className="observation-modal__student">
              <Space direction="vertical" size={4}>
                <Typography.Text strong>
                  {quickVisit.student_name} / {quickVisit.class_name}
                </Typography.Text>
                <Typography.Text type="secondary">
                  {quickVisit.symptoms.map(symptomLabel).join("、") || "暂无主诉"} ·{" "}
                  {formatObservationDuration(quickVisit.created_at)}
                </Typography.Text>
              </Space>
              <Tag color="blue">当前留观</Tag>
            </div>
            <Segmented
              block
              value={quickDestination}
              onChange={(value) => setQuickDestination(String(value))}
              options={[
                { label: "继续留观", value: "observation" },
                { label: "返回班级", value: "return_class" },
                { label: "转诊", value: "hospital" },
                { label: "紧急处理", value: "urgent" },
              ]}
            />
            {quickDestination !== "observation" ? (
              <Alert
                type="info"
                showIcon
                message={`保存后该学生会从“当前留观人员”中移出，并同步到就诊记录。`}
              />
            ) : null}
          </Space>
        ) : null}
      </Modal>

      <div className="visits-workbench__header">
        <div>
          <Typography.Title level={3} style={{ marginBottom: 4 }}>
            就诊队列
          </Typography.Title>
          <Typography.Text type="secondary">
            学生登记、体温计倒计时和就诊记录会通过 Socket 自动同步。
          </Typography.Text>
        </div>
        <Space wrap>
          <Tag color={observationVisits.length ? "blue" : "green"}>
            留观 {observationVisits.length} 人
          </Tag>
          <Tag color="orange">使用中 {usedThermometers.size} 根</Tag>
          <Tag color="green">空闲 {idleThermometerCount} 根</Tag>
        </Space>
      </div>

      <div className="visits-workbench__grid">
        <Card className="doctor-live-card" title="实时学生登记">
          {liveStudentList.length ? (
            <Space direction="vertical" size={12} style={{ width: "100%" }}>
              {liveStudentList.map((student) => (
                <div className="doctor-live-card__item" key={student.session_id}>
                  <div className="doctor-live-card__row">
                    <Space size={10}>
                      <IdcardOutlined />
                      <Typography.Text strong>
                        {student.student_name || student.student_id || "待识别学生"}
                      </Typography.Text>
                      {student.student_id ? <Tag>{student.student_id}</Tag> : null}
                    </Space>
                    <Tag color={student.status === "temperature" ? "gold" : "blue"}>
                      {student.status === "temperature" ? "体温流程" : "登记中"}
                    </Tag>
                  </div>
                  <Typography.Paragraph className="doctor-live-card__question">
                    {student.question}
                  </Typography.Paragraph>
                  <Space wrap>
                    {(student.symptoms ?? []).length ? (
                      student.symptoms.map((symptom) => (
                        <Tag key={symptom}>{symptomLabel(symptom)}</Tag>
                      ))
                    ) : (
                      <Tag color="green">无明显主诉</Tag>
                    )}
                    <Tag color={temperatureColor(student.temperature_status ?? "normal")}>
                      {student.temperature_status === "measured" && student.temperature
                        ? `${student.temperature.toFixed(1)}℃`
                        : student.temperature_status === "timing"
                          ? `计时 ${formatCountdown(student.countdown_seconds)}`
                          : student.temperature_status === "due"
                            ? "到时待录入"
                            : "体温正常"}
                    </Tag>
                  </Space>
                  <div className="doctor-live-card__actions">
                    <Tooltip title="仅在医生判断需要时使用水银体温计">
                      <Button
                        disabled={isActiveThermometer(student.temperature_status)}
                        icon={<ExperimentOutlined />}
                        onClick={() => requestTemperature(student)}
                      >
                        要求测温
                      </Button>
                    </Tooltip>
                  </div>
                </div>
              ))}
            </Space>
          ) : (
            <div className="doctor-live-card__empty">
              <Empty description="暂无学生正在登记" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            </div>
          )}
        </Card>

        <Card
          className="doctor-thermometer-card"
          title={
            <Space>
              <ExperimentOutlined />
              <span>水银体温计管理</span>
            </Space>
          }
        >
          <div className="thermometer-rack">
            <div className="thermometer-rack__metric is-busy">
              <strong>{usedThermometers.size}</strong>
              <span>使用中</span>
            </div>
            <div className="thermometer-rack__metric is-idle">
              <strong>{idleThermometerCount}</strong>
              <span>空闲</span>
            </div>
          </div>

          {activeThermometerEntries.length ? (
            <Space direction="vertical" size={12} style={{ width: "100%" }}>
              {activeThermometerEntries.map((entry) => (
                <div
                  className={`thermometer-entry ${
                    entry.temperature_status === "due" ? "is-due" : ""
                  }`}
                  key={entry.session_id}
                >
                  <div className="thermometer-entry__main">
                    <Space size={10}>
                      {entry.temperature_status === "due" ? (
                        <FireOutlined />
                      ) : (
                        <ClockCircleOutlined />
                      )}
                      <Typography.Text strong>
                        体温计 {entry.thermometer_id} · {entry.student_name || entry.student_id}
                      </Typography.Text>
                    </Space>
                    <span className="doctor-live-card__timer">
                      {formatCountdown(entry.countdown_seconds)}
                    </span>
                  </div>
                  {entry.temperature_status === "due" ? (
                    <Alert type="warning" showIcon message="已到 5 分钟，请读数后录入体温。" />
                  ) : null}
                  <div className="thermometer-entry__input">
                    <InputNumber
                      min={34}
                      max={43}
                      step={0.1}
                      precision={1}
                      placeholder="体温"
                      value={temperatureInputs[entry.session_id] ?? null}
                      onChange={(value) =>
                        setTemperatureInputs((current) => ({
                          ...current,
                          [entry.session_id]: typeof value === "number" ? value : null,
                        }))
                      }
                    />
                    <Button
                      icon={<CheckCircleOutlined />}
                      type={entry.temperature_status === "due" ? "primary" : "default"}
                      onClick={() => recordTemperature(entry)}
                    >
                      录入并同步
                    </Button>
                  </div>
                </div>
              ))}
            </Space>
          ) : (
            <div className="doctor-live-card__empty">
              <Empty description="暂无体温计在使用" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            </div>
          )}
        </Card>
      </div>

      <Card
        className="observation-board"
        title={
          <Space>
            <UserSwitchOutlined />
            <span>医务室目前留观人员</span>
            <Badge count={observationVisits.length} overflowCount={99} />
          </Space>
        }
        extra={
          observationWarningCount > 0 ? (
            <Tag color="orange">
              <AlertOutlined /> {observationWarningCount} 人需优先复核
            </Tag>
          ) : (
            <Tag color="green">状态清晰</Tag>
          )
        }
      >
        {observationVisits.length ? (
          <div className="observation-board__grid">
            {observationVisits.map((visit) => {
              const elevated = (visit.temperature_value ?? 0) >= 37.3;
              return (
                <button
                  className={`observation-card ${elevated ? "is-warning" : ""}`}
                  key={visit.id}
                  type="button"
                  onClick={() => openQuickDestination(visit)}
                >
                  <div className="observation-card__topline">
                    <span>{visit.student_name}</span>
                    <Tag color={elevated ? "orange" : "blue"}>留观中</Tag>
                  </div>
                  <div className="observation-card__meta">
                    <span>{visit.class_name}</span>
                    <span>{formatObservationDuration(visit.created_at)}</span>
                  </div>
                  <div className="observation-card__symptoms">
                    {(visit.symptoms ?? []).slice(0, 3).map((symptom) => (
                      <Tag key={symptom}>{symptomLabel(symptom)}</Tag>
                    ))}
                    {(visit.symptoms ?? []).length > 3 ? (
                      <Tag>+{visit.symptoms.length - 3}</Tag>
                    ) : null}
                    {visit.temperature_status === "measured" && visit.temperature_value !== null ? (
                      <Tag color={elevated ? "orange" : "green"}>
                        {visit.temperature_value.toFixed(1)}℃
                      </Tag>
                    ) : null}
                  </div>
                </button>
              );
            })}
          </div>
        ) : (
          <div className="doctor-live-card__empty">
            <Empty description="当前没有留观人员" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          </div>
        )}
      </Card>

      <Card title="已提交就诊记录">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={visits.map(toRow)}
          loading={loading}
          rowClassName={(row) => (row.isObservation ? "visit-row-observation" : "")}
          pagination={{ pageSize: 10, showSizeChanger: false }}
        />
      </Card>
    </Space>
  );
}
