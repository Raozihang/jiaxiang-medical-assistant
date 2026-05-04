import {
  Alert,
  Button,
  Card,
  Col,
  DatePicker,
  Descriptions,
  Divider,
  Form,
  Input,
  List,
  message,
  Row,
  Select,
  Space,
  Tag,
  Typography,
} from "antd";
import dayjs, { type Dayjs } from "dayjs";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  type AnalyzeResult,
  checkMedicineInteractions,
  type InteractionCheckResult,
  parseAnalyzeResult,
  parseInteractionResult,
  parseRecommendResult,
  parseTriageResult,
  type RecommendResult,
  type TriageResult,
} from "@/shared/api/ai";
import { getErrorMessage } from "@/shared/api/helpers";
import { getVisit, regenerateVisitAIAnalysis, updateVisit, type Visit } from "@/shared/api/visits";
import {
  getDestinationLabel,
  getSymptomLabel,
  normalizeDestinationForForm,
} from "@/shared/labels/localization";

function symptomLabel(value: string) {
  return getSymptomLabel(value);
}

type UpdateForm = {
  diagnosis: string;
  prescription: string;
  destination: string;
  follow_up_at: Dayjs | null;
  follow_up_note: string;
};

function parsePrescriptionInput(value: string) {
  return value
    .split(/[,\n;|]/g)
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
}

function normalizeDestination(value: string | null | undefined) {
  return normalizeDestinationForForm(value);
}

function destinationLabel(value: string) {
  return getDestinationLabel(value);
}

function destinationTagColor(value: string) {
  const normalized = normalizeDestination(value);
  if (normalized === "urgent") {
    return "red";
  }
  if (normalized === "hospital") {
    return "orange";
  }
  if (normalized === "return_class") {
    return "green";
  }
  return "blue";
}

function parseFollowUpAt(value: string | null | undefined) {
  if (!value) {
    return null;
  }
  const parsed = dayjs(value);
  return parsed.isValid() ? parsed : null;
}

function formatFollowUpAt(value: string | null | undefined) {
  const parsed = parseFollowUpAt(value);
  return parsed ? parsed.format("YYYY-MM-DD HH:mm") : "-";
}

function isAIAnalysisRunning(status: string | undefined) {
  return status === "queued" || status === "processing";
}

function aiAnalysisStatusText(status: string | undefined) {
  switch (status) {
    case "queued":
      return "AI 建议已进入后台队列";
    case "processing":
      return "AI 建议生成中";
    case "completed":
      return "AI 建议已生成";
    case "failed":
      return "AI 建议生成失败";
    default:
      return "AI 建议尚未生成";
  }
}

function aiAnalysisStatusColor(status: string | undefined) {
  switch (status) {
    case "queued":
    case "processing":
      return "processing";
    case "completed":
      return "success";
    case "failed":
      return "error";
    default:
      return "default";
  }
}

function levelColor(level: string) {
  if (["critical", "high", "severe"].includes(level.toLowerCase())) {
    return "red";
  }
  if (["medium", "warning"].includes(level.toLowerCase())) {
    return "orange";
  }
  return "blue";
}

export function VisitDetailPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [regenerating, setRegenerating] = useState(false);
  const [checkingInteraction, setCheckingInteraction] = useState(false);
  const [visit, setVisit] = useState<Visit | null>(null);
  const [analyzeResult, setAnalyzeResult] = useState<AnalyzeResult | null>(null);
  const [triageResult, setTriageResult] = useState<TriageResult | null>(null);
  const [recommendResult, setRecommendResult] = useState<RecommendResult | null>(null);
  const [interactionResult, setInteractionResult] = useState<InteractionCheckResult | null>(null);
  const [form] = Form.useForm<UpdateForm>();
  const [messageApi, contextHolder] = message.useMessage();

  const currentPrescription = Form.useWatch("prescription", form);
  const aiStatus = visit?.ai_analysis?.status ?? "not_started";
  const aiRunning = isAIAnalysisRunning(aiStatus);

  const currentMedicines = useMemo(() => {
    if (!currentPrescription) {
      return [];
    }
    return parsePrescriptionInput(currentPrescription);
  }, [currentPrescription]);

  const applyCachedAIAnalysis = useCallback((data: Visit) => {
    const snapshot = data.ai_analysis;
    setAnalyzeResult(snapshot?.analyze ? parseAnalyzeResult(snapshot.analyze) : null);
    setTriageResult(snapshot?.triage ? parseTriageResult(snapshot.triage) : null);
    setRecommendResult(snapshot?.recommend ? parseRecommendResult(snapshot.recommend) : null);
    setInteractionResult(
      snapshot?.interaction ? parseInteractionResult(snapshot.interaction) : null,
    );
  }, []);

  const hydrateVisit = useCallback(
    (data: Visit, syncForm: boolean) => {
      setVisit(data);
      applyCachedAIAnalysis(data);
      if (syncForm) {
        form.setFieldsValue({
          diagnosis: data.diagnosis ?? "",
          prescription: (data.prescription ?? []).join(", "),
          destination: normalizeDestination(data.destination),
          follow_up_at: parseFollowUpAt(data.follow_up_at),
          follow_up_note: data.follow_up_note ?? "",
        });
      }
    },
    [applyCachedAIAnalysis, form],
  );

  const loadDetail = useCallback(async () => {
    if (!id) {
      return;
    }

    setLoading(true);
    try {
      const data = await getVisit(id);
      hydrateVisit(data, true);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "就诊详情加载失败"));
    } finally {
      setLoading(false);
    }
  }, [hydrateVisit, id, messageApi]);

  useEffect(() => {
    void loadDetail();
  }, [loadDetail]);

  useEffect(() => {
    if (!id || !aiRunning) {
      return;
    }

    const timer = window.setInterval(async () => {
      try {
        const data = await getVisit(id);
        hydrateVisit(data, false);
      } catch {
        window.clearInterval(timer);
      }
    }, 2500);

    return () => window.clearInterval(timer);
  }, [aiRunning, hydrateVisit, id]);

  const regenerateAIAnalysis = async () => {
    if (!visit) {
      return;
    }
    setRegenerating(true);
    try {
      const data = await regenerateVisitAIAnalysis(visit.id);
      hydrateVisit(data, false);
      messageApi.success("AI 建议已进入后台队列");
    } catch (error) {
      messageApi.error(getErrorMessage(error, "AI 建议重新生成失败"));
    } finally {
      setRegenerating(false);
    }
  };

  const runInteractionCheck = async () => {
    if (!visit) {
      return;
    }

    const medicines = currentMedicines.length
      ? currentMedicines
      : (recommendResult?.medicines.map((item) => item.name) ?? []);

    if (!medicines.length) {
      messageApi.warning("请先填写处方，或等待后台生成药品建议");
      return;
    }

    setCheckingInteraction(true);
    try {
      const result = await checkMedicineInteractions({
        medicines,
        student_id: visit.student_id,
      });
      setInteractionResult(result);
      if (result.hasInteraction) {
        messageApi.warning("发现潜在药物相互作用");
      } else {
        messageApi.success("未发现明显相互作用");
      }
    } catch (error) {
      messageApi.error(getErrorMessage(error, "相互作用检查失败"));
    } finally {
      setCheckingInteraction(false);
    }
  };

  const applyRecommendedMedicines = () => {
    if (!recommendResult?.medicines.length) {
      return;
    }
    form.setFieldValue(
      "prescription",
      recommendResult.medicines.map((item) => item.name).join(", "),
    );
    messageApi.success("已填充 AI 推荐药品，请医生确认后保存");
  };

  const applyTriageDestination = () => {
    if (!triageResult?.destination) {
      return;
    }
    form.setFieldValue("destination", normalizeDestination(triageResult.destination));
    messageApi.success("已应用 AI 建议去向，请医生确认后保存");
  };

  const handleSave = async (values: UpdateForm) => {
    if (!id) {
      return;
    }

    setSaving(true);
    try {
      await updateVisit(id, {
        diagnosis: values.diagnosis,
        prescription: parsePrescriptionInput(values.prescription),
        destination: values.destination,
        follow_up_at: values.follow_up_at ? values.follow_up_at.toDate().toISOString() : "",
        follow_up_note: values.follow_up_note.trim(),
      });
      messageApi.success("就诊记录已更新");
      await loadDetail();
    } catch (error) {
      messageApi.error(getErrorMessage(error, "保存失败"));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        就诊详情与 AI 辅助建议
      </Typography.Title>

      <Card loading={loading}>
        {visit ? (
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <Descriptions bordered size="small" column={2}>
              <Descriptions.Item label="就诊 ID">{visit.id}</Descriptions.Item>
              <Descriptions.Item label="学生信息">
                {visit.student_name} / {visit.class_name}
              </Descriptions.Item>
              <Descriptions.Item label="症状" span={2}>
                {(visit.symptoms ?? []).map(symptomLabel).join("、") || "-"}
              </Descriptions.Item>
              <Descriptions.Item label="主诉" span={2}>
                {visit.description || "-"}
              </Descriptions.Item>
              <Descriptions.Item label="当前去向">
                <Tag color={destinationTagColor(visit.destination)}>
                  {destinationLabel(visit.destination)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="复诊提醒">
                {formatFollowUpAt(visit.follow_up_at)}
              </Descriptions.Item>
              <Descriptions.Item label="复诊备注" span={2}>
                {visit.follow_up_note?.trim() || "-"}
              </Descriptions.Item>
            </Descriptions>

            <Form layout="vertical" form={form} onFinish={(values) => void handleSave(values)}>
              <Form.Item label="诊断结果" name="diagnosis">
                <Input.TextArea rows={3} placeholder="可根据 AI 建议调整后录入" />
              </Form.Item>
              <Form.Item label="处方（逗号或换行分隔）" name="prescription">
                <Input.TextArea rows={4} placeholder="例：布洛芬，氯雷他定" />
              </Form.Item>
              <Form.Item label="去向" name="destination">
                <Select
                  options={[
                    { label: "留观", value: "observation" },
                    { label: "返回班级", value: "return_class" },
                    { label: "转诊", value: "hospital" },
                    { label: "紧急处理", value: "urgent" },
                  ]}
                />
              </Form.Item>
              <Form.Item label="复诊时间" name="follow_up_at">
                <DatePicker
                  showTime={{ format: "HH:mm" }}
                  format="YYYY-MM-DD HH:mm"
                  placeholder="请选择复诊时间"
                  style={{ width: "100%" }}
                />
              </Form.Item>
              <Form.Item label="复诊备注" name="follow_up_note">
                <Input.TextArea rows={3} placeholder="可填写复诊关注点或提醒说明" />
              </Form.Item>
              <Button htmlType="submit" type="primary" loading={saving}>
                保存就诊记录
              </Button>
            </Form>
          </Space>
        ) : (
          <Typography.Text type="secondary">暂无就诊数据</Typography.Text>
        )}
      </Card>

      <Card
        title="AI 辅助建议"
        extra={
          <Space wrap>
            <Tag color={aiAnalysisStatusColor(aiStatus)}>{aiAnalysisStatusText(aiStatus)}</Tag>
            {visit?.ai_analysis?.processed_at ? (
              <Typography.Text type="secondary">
                {dayjs(visit.ai_analysis.processed_at).format("MM-DD HH:mm")}
              </Typography.Text>
            ) : null}
            <Button
              type="primary"
              loading={regenerating || aiRunning}
              onClick={() => void regenerateAIAnalysis()}
            >
              重新生成 AI 建议
            </Button>
          </Space>
        }
      >
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          {visit?.ai_analysis?.error ? (
            <Alert
              type="error"
              showIcon
              message="后台 AI 生成失败"
              description={visit.ai_analysis.error}
            />
          ) : null}
          {aiRunning ? (
            <Alert
              type="info"
              showIcon
              message="后台正在生成 AI 建议"
              description="医生可以继续查看或编辑就诊记录，生成完成后此面板会自动刷新缓存结果。"
            />
          ) : null}

          <Row gutter={[16, 16]}>
            <Col span={12}>
              <Card title="症状结构化" type="inner" loading={aiRunning && !analyzeResult}>
                {analyzeResult ? (
                  <Space direction="vertical" size={12} style={{ width: "100%" }}>
                    <Typography.Paragraph>{analyzeResult.summary}</Typography.Paragraph>
                    <List
                      size="small"
                      dataSource={analyzeResult.structuredSymptoms}
                      locale={{ emptyText: "暂无结构化症状" }}
                      renderItem={(item) => (
                        <List.Item>
                          <Space direction="vertical" size={0}>
                            <Typography.Text strong>{item.name}</Typography.Text>
                            <Typography.Text type="secondary">
                              严重程度：{item.severity ?? "-"} | 持续时间：{item.duration ?? "-"}
                            </Typography.Text>
                            {item.note ? (
                              <Typography.Text type="secondary">备注：{item.note}</Typography.Text>
                            ) : null}
                          </Space>
                        </List.Item>
                      )}
                    />
                    {analyzeResult.possibleConditions.length ? (
                      <>
                        <Divider style={{ margin: "8px 0" }} />
                        <Typography.Text strong>可能诊断：</Typography.Text>
                        <Space wrap>
                          {analyzeResult.possibleConditions.map((item) => (
                            <Tag key={item}>{item}</Tag>
                          ))}
                        </Space>
                      </>
                    ) : null}
                  </Space>
                ) : (
                  <Typography.Text type="secondary">
                    暂无缓存结果，可点击“重新生成 AI 建议”入队生成
                  </Typography.Text>
                )}
              </Card>
            </Col>

            <Col span={12}>
              <Card
                title="分诊建议"
                type="inner"
                loading={aiRunning && !triageResult}
                extra={
                  <Button size="small" onClick={applyTriageDestination} disabled={!triageResult}>
                    应用去向
                  </Button>
                }
              >
                {triageResult ? (
                  <Space direction="vertical" size={10} style={{ width: "100%" }}>
                    <Typography.Text>
                      分诊等级：<Tag>{triageResult.level}</Tag>
                    </Typography.Text>
                    <Typography.Text>
                      建议去向：
                      <Tag
                        color={destinationTagColor(normalizeDestination(triageResult.destination))}
                      >
                        {destinationLabel(normalizeDestination(triageResult.destination))}
                      </Tag>
                    </Typography.Text>
                    <Typography.Paragraph>{triageResult.reason}</Typography.Paragraph>
                    <List
                      size="small"
                      header="分诊建议"
                      dataSource={triageResult.recommendations}
                      locale={{ emptyText: "暂无分诊建议" }}
                      renderItem={(item) => <List.Item>{item}</List.Item>}
                    />
                    {triageResult.riskFlags.length ? (
                      <Alert
                        type="warning"
                        showIcon
                        message="风险提示"
                        description={triageResult.riskFlags.join("；")}
                      />
                    ) : null}
                  </Space>
                ) : (
                  <Typography.Text type="secondary">暂无缓存分诊建议</Typography.Text>
                )}
              </Card>
            </Col>

            <Col span={12}>
              <Card
                title="药品建议"
                type="inner"
                loading={aiRunning && !recommendResult}
                extra={
                  <Button
                    size="small"
                    onClick={applyRecommendedMedicines}
                    disabled={!recommendResult?.medicines.length}
                  >
                    应用到处方
                  </Button>
                }
              >
                {recommendResult ? (
                  <Space direction="vertical" size={10} style={{ width: "100%" }}>
                    <List
                      size="small"
                      dataSource={recommendResult.medicines}
                      locale={{ emptyText: "暂无推荐药品" }}
                      renderItem={(item) => (
                        <List.Item>
                          <Space direction="vertical" size={0}>
                            <Typography.Text strong>{item.name}</Typography.Text>
                            <Typography.Text type="secondary">
                              用法：{item.dosage || "-"} {item.frequency || ""}{" "}
                              {item.duration || ""}
                            </Typography.Text>
                            {item.reason ? (
                              <Typography.Text type="secondary">
                                原因：{item.reason}
                              </Typography.Text>
                            ) : null}
                            {item.caution ? (
                              <Typography.Text type="warning">注意：{item.caution}</Typography.Text>
                            ) : null}
                          </Space>
                        </List.Item>
                      )}
                    />
                    {recommendResult.contraindications.length ? (
                      <Alert
                        type="warning"
                        showIcon
                        message="禁忌提示"
                        description={recommendResult.contraindications.join("；")}
                      />
                    ) : null}
                  </Space>
                ) : (
                  <Typography.Text type="secondary">暂无缓存药品建议</Typography.Text>
                )}
              </Card>
            </Col>

            <Col span={12}>
              <Card
                title="药物相互作用检查"
                type="inner"
                loading={(aiRunning && !interactionResult) || checkingInteraction}
                extra={
                  <Button size="small" onClick={() => void runInteractionCheck()} disabled={!visit}>
                    检查当前处方
                  </Button>
                }
              >
                {interactionResult ? (
                  <Space direction="vertical" size={10} style={{ width: "100%" }}>
                    <Alert
                      type={interactionResult.hasInteraction ? "error" : "success"}
                      showIcon
                      message={
                        interactionResult.hasInteraction
                          ? `检测到相互作用（等级：${interactionResult.severity}）`
                          : "未发现明显相互作用"
                      }
                      description={
                        interactionResult.hasInteraction
                          ? "建议复核推荐处方并结合病情调整"
                          : "仍建议结合学生过敏史与临床经验判断"
                      }
                    />
                    <List
                      size="small"
                      dataSource={interactionResult.warnings}
                      locale={{ emptyText: "暂无冲突明细" }}
                      renderItem={(item) => (
                        <List.Item>
                          <Space direction="vertical" size={0}>
                            <Typography.Text strong>
                              {item.title}{" "}
                              <Tag color={levelColor(item.severity)}>{item.severity}</Tag>
                            </Typography.Text>
                            <Typography.Text>{item.description}</Typography.Text>
                            <Typography.Text type="secondary">
                              建议：{item.suggestion}
                            </Typography.Text>
                          </Space>
                        </List.Item>
                      )}
                    />
                  </Space>
                ) : (
                  <Typography.Text type="secondary">暂无缓存检查结果</Typography.Text>
                )}
              </Card>
            </Col>
          </Row>
        </Space>
      </Card>
    </Space>
  );
}
