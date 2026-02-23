import {
  Alert,
  Button,
  Card,
  Col,
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
import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  analyzeSymptoms,
  checkMedicineInteractions,
  recommendMedicines,
  triageVisit,
  type AnalyzeResult,
  type InteractionCheckResult,
  type RecommendResult,
  type TriageResult,
} from "@/shared/api/ai";
import { getErrorMessage } from "@/shared/api/helpers";
import { getVisit, updateVisit, type Visit } from "@/shared/api/visits";

type UpdateForm = {
  diagnosis: string;
  prescription: string;
  destination: string;
};

type AILoadingState = {
  analyze: boolean;
  triage: boolean;
  recommend: boolean;
  interaction: boolean;
};

const defaultAILoading: AILoadingState = {
  analyze: false,
  triage: false,
  recommend: false,
  interaction: false,
};

function parsePrescriptionInput(value: string) {
  return value
    .split(/[,\n;|]/g)
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
}

function normalizeDestination(value: string) {
  const normalized = value.trim().toLowerCase();
  if (["urgent", "critical", "high"].includes(normalized)) {
    return "urgent";
  }
  if (["hospital", "referral", "transfer"].includes(normalized)) {
    return "hospital";
  }
  if (["return_class", "returnclass", "class", "back"].includes(normalized)) {
    return "return_class";
  }
  return "observation";
}

function destinationLabel(value: string) {
  switch (value) {
    case "urgent":
      return "Urgent";
    case "hospital":
      return "Hospital Referral";
    case "return_class":
      return "Return Class";
    default:
      return "Observation";
  }
}

function destinationTagColor(value: string) {
  if (value === "urgent") {
    return "red";
  }
  if (value === "hospital") {
    return "orange";
  }
  if (value === "return_class") {
    return "green";
  }
  return "blue";
}

export function VisitDetailPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [visit, setVisit] = useState<Visit | null>(null);
  const [analyzeResult, setAnalyzeResult] = useState<AnalyzeResult | null>(null);
  const [triageResult, setTriageResult] = useState<TriageResult | null>(null);
  const [recommendResult, setRecommendResult] = useState<RecommendResult | null>(null);
  const [interactionResult, setInteractionResult] = useState<InteractionCheckResult | null>(null);
  const [aiLoading, setAiLoading] = useState<AILoadingState>(defaultAILoading);
  const [form] = Form.useForm<UpdateForm>();
  const [messageApi, contextHolder] = message.useMessage();

  const currentPrescription = Form.useWatch("prescription", form);

  const currentMedicines = useMemo(() => {
    if (!currentPrescription) {
      return [];
    }
    return parsePrescriptionInput(currentPrescription);
  }, [currentPrescription]);

  const setSingleAILoading = useCallback((key: keyof AILoadingState, value: boolean) => {
    setAiLoading((prev) => ({ ...prev, [key]: value }));
  }, []);

  const loadDetail = useCallback(async () => {
    if (!id) {
      return;
    }

    setLoading(true);
    try {
      const data = await getVisit(id);
      setVisit(data);
      form.setFieldsValue({
        diagnosis: data.diagnosis ?? "",
        prescription: data.prescription.join(", "),
        destination: data.destination || "observation",
      });
      setAnalyzeResult(null);
      setTriageResult(null);
      setRecommendResult(null);
      setInteractionResult(null);
    } catch (error) {
      messageApi.error(getErrorMessage(error, "就诊详情加载失败"));
    } finally {
      setLoading(false);
    }
  }, [form, id, messageApi]);

  useEffect(() => {
    void loadDetail();
  }, [loadDetail]);

  const runAnalyze = async () => {
    if (!visit) {
      return;
    }
    setSingleAILoading("analyze", true);
    try {
      const result = await analyzeSymptoms({
        visit_id: visit.id,
        symptoms: visit.symptoms,
        description: visit.description,
      });
      setAnalyzeResult(result);
      messageApi.success("症状结构化完成");
    } catch (error) {
      messageApi.error(getErrorMessage(error, "症状结构化失败"));
    } finally {
      setSingleAILoading("analyze", false);
    }
  };

  const runTriage = async () => {
    if (!visit) {
      return;
    }
    setSingleAILoading("triage", true);
    try {
      const result = await triageVisit({
        visit_id: visit.id,
        symptoms: visit.symptoms,
        description: visit.description,
        analysis_summary: analyzeResult?.summary,
      });
      setTriageResult(result);
      messageApi.success("智能分诊完成");
    } catch (error) {
      messageApi.error(getErrorMessage(error, "智能分诊失败"));
    } finally {
      setSingleAILoading("triage", false);
    }
  };

  const runRecommend = async () => {
    if (!visit) {
      return;
    }
    setSingleAILoading("recommend", true);
    try {
      const diagnosis = form.getFieldValue("diagnosis");
      const result = await recommendMedicines({
        visit_id: visit.id,
        symptoms: visit.symptoms,
        diagnosis,
        triage_level: triageResult?.level,
      });
      setRecommendResult(result);
      messageApi.success("药品推荐完成");
    } catch (error) {
      messageApi.error(getErrorMessage(error, "药品推荐失败"));
    } finally {
      setSingleAILoading("recommend", false);
    }
  };

  const runInteractionCheck = async () => {
    if (!visit) {
      return;
    }

    const medicines = currentMedicines.length
      ? currentMedicines
      : recommendResult?.medicines.map((item) => item.name) ?? [];

    if (!medicines.length) {
      messageApi.warning("请先填写处方或先执行药品推荐");
      return;
    }

    setSingleAILoading("interaction", true);
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
      setSingleAILoading("interaction", false);
    }
  };

  const runAllAI = async () => {
    await runAnalyze();
    await runTriage();
    await runRecommend();
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
    messageApi.success("已应用 AI 分诊去向，请医生确认后保存");
  };

  const handleSave = async (values: UpdateForm) => {
    if (!id) {
      return;
    }

    const prescription = parsePrescriptionInput(values.prescription);

    setSaving(true);
    try {
      await updateVisit(id, {
        diagnosis: values.diagnosis,
        prescription,
        destination: values.destination,
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
        就诊详情与 AI 辅助
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
                {visit.symptoms.join(", ") || "-"}
              </Descriptions.Item>
              <Descriptions.Item label="主诉" span={2}>
                {visit.description || "-"}
              </Descriptions.Item>
              <Descriptions.Item label="当前去向">
                <Tag color={destinationTagColor(visit.destination)}>{destinationLabel(visit.destination)}</Tag>
              </Descriptions.Item>
            </Descriptions>

            <Form layout="vertical" form={form} onFinish={(values) => void handleSave(values)}>
              <Form.Item label="诊断结果" name="diagnosis">
                <Input.TextArea rows={3} placeholder="可根据 AI 建议调整后录入" />
              </Form.Item>
              <Form.Item label="处方（逗号或换行分隔）" name="prescription">
                <Input.TextArea rows={4} placeholder="例：布洛芬, 氯雷他定" />
              </Form.Item>
              <Form.Item label="去向" name="destination">
                <Select
                  options={[
                    { label: "Observation", value: "observation" },
                    { label: "Return Class", value: "return_class" },
                    { label: "Hospital Referral", value: "hospital" },
                    { label: "Urgent", value: "urgent" },
                  ]}
                />
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
        title="AI 辅助决策"
        extra={
          <Space>
            <Button loading={aiLoading.analyze} onClick={() => void runAnalyze()}>
              症状结构化
            </Button>
            <Button loading={aiLoading.triage} onClick={() => void runTriage()}>
              智能分诊
            </Button>
            <Button loading={aiLoading.recommend} onClick={() => void runRecommend()}>
              药品推荐
            </Button>
            <Button loading={aiLoading.interaction} onClick={() => void runInteractionCheck()}>
              相互作用检查
            </Button>
            <Button type="primary" onClick={() => void runAllAI()}>
              一键生成建议
            </Button>
          </Space>
        }
      >
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <Row gutter={[16, 16]}>
            <Col span={12}>
              <Card title="症状结构化" type="inner" loading={aiLoading.analyze}>
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
                  <Typography.Text type="secondary">点击“症状结构化”生成 AI 建议</Typography.Text>
                )}
              </Card>
            </Col>

            <Col span={12}>
              <Card
                title="智能分诊"
                type="inner"
                loading={aiLoading.triage}
                extra={
                  <Button size="small" onClick={applyTriageDestination} disabled={!triageResult}>
                    应用去向
                  </Button>
                }
              >
                {triageResult ? (
                  <Space direction="vertical" size={10} style={{ width: "100%" }}>
                    <Typography.Text>
                      分诊等级：<Tag color={destinationTagColor(normalizeDestination(triageResult.destination))}>{triageResult.level}</Tag>
                    </Typography.Text>
                    <Typography.Text>
                      建议去向：
                      <Tag color={destinationTagColor(normalizeDestination(triageResult.destination))}>
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
                  <Typography.Text type="secondary">点击“智能分诊”获取 AI 去向建议</Typography.Text>
                )}
              </Card>
            </Col>

            <Col span={12}>
              <Card
                title="药品推荐"
                type="inner"
                loading={aiLoading.recommend}
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
                              用法：{item.dosage || "-"} {item.frequency || ""} {item.duration || ""}
                            </Typography.Text>
                            {item.reason ? (
                              <Typography.Text type="secondary">原因：{item.reason}</Typography.Text>
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
                  <Typography.Text type="secondary">点击“药品推荐”生成候选方案</Typography.Text>
                )}
              </Card>
            </Col>

            <Col span={12}>
              <Card title="药物相互作用检查" type="inner" loading={aiLoading.interaction}>
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
                              {item.title} <Tag color={levelColor(item.severity)}>{item.severity}</Tag>
                            </Typography.Text>
                            <Typography.Text>{item.description}</Typography.Text>
                            <Typography.Text type="secondary">建议：{item.suggestion}</Typography.Text>
                          </Space>
                        </List.Item>
                      )}
                    />
                  </Space>
                ) : (
                  <Typography.Text type="secondary">点击“相互作用检查”审核当前处方</Typography.Text>
                )}
              </Card>
            </Col>
          </Row>
        </Space>
      </Card>
    </Space>
  );
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

