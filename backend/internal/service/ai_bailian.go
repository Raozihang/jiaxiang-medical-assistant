package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type bailianProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewBailianProvider(apiKey, model, baseURL string) AIProvider {
	return &bailianProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// OpenAI-compatible request/response types for DashScope.

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *chatError   `json:"error,omitempty"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (p *bailianProvider) call(ctx context.Context, system, user string) (string, error) {
	body := chatRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.3,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("dashscope request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dashscope returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("dashscope error [%s]: %s", chatResp.Error.Code, chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("dashscope returned no choices")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// extractJSON finds the first JSON object or array in the text, handling
// markdown code fences that LLMs sometimes wrap around their output.
func extractJSON(text string) string {
	s := strings.TrimSpace(text)
	// Strip markdown code fence if present.
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}

// ---- Analyze ----

const analyzeSystemPrompt = `你是一个校园医务室的 AI 辅助分析系统。根据学生的症状和体温，输出 JSON 格式的分析结果。

要求：
1. risk_level: "low"、"medium" 或 "high"
2. confidence: 0.0-1.0 之间的置信度
3. matched_signals: 从症状中匹配到的关键信号列表
4. possible_causes: 可能的病因列表（中文）
5. suggested_actions: 建议措施列表（中文）

仅输出 JSON，不要输出其他内容。示例格式：
{"risk_level":"medium","confidence":0.75,"matched_signals":["发烧","头痛"],"possible_causes":["上呼吸道感染"],"suggested_actions":["留观并每30分钟测量体温"]}`

func (p *bailianProvider) Analyze(ctx context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	userMsg := fmt.Sprintf("症状列表: %s\n描述: %s\n体温: %.1f°C",
		strings.Join(input.Symptoms, ", "), input.Description, input.Temperature)

	raw, err := p.call(ctx, analyzeSystemPrompt, userMsg)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("AI analyze: %w", err)
	}

	var result AnalyzeResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return AnalyzeResult{}, fmt.Errorf("parse AI analyze response: %w (raw: %s)", err, raw)
	}
	return result, nil
}

// ---- Triage ----

const triageSystemPrompt = `你是一个校园医务室的 AI 智能分诊系统。根据学生的症状和体温，判断分诊级别和去向。

要求：
1. triage_level: "routine"（普通）、"priority"（优先）或 "urgent"（紧急）
2. destination: "classroom"（返回教室）、"observation"（留观）或 "hospital"（转院）
3. reason: 分诊理由（中文）
4. review_in_minutes: 建议复查间隔（分钟数）
5. suggested_actions: 建议措施列表（中文）

紧急信号包括：胸痛、呼吸困难、抽搐、晕厥、呕血、体温≥39.5°C。
发现紧急信号时必须判定为 urgent 并建议转院。

仅输出 JSON，不要输出其他内容。`

func (p *bailianProvider) Triage(ctx context.Context, input TriageInput) (TriageResult, error) {
	userMsg := fmt.Sprintf("症状列表: %s\n描述: %s\n体温: %.1f°C",
		strings.Join(input.Symptoms, ", "), input.Description, input.Temperature)

	raw, err := p.call(ctx, triageSystemPrompt, userMsg)
	if err != nil {
		return TriageResult{}, fmt.Errorf("AI triage: %w", err)
	}

	var result TriageResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return TriageResult{}, fmt.Errorf("parse AI triage response: %w (raw: %s)", err, raw)
	}
	return result, nil
}

// ---- Recommend ----

const recommendSystemPrompt = `你是一个校园医务室的 AI 诊疗建议系统。根据诊断结果和症状，给出护理计划和药品建议。

要求：
1. plan_version: 固定为 "ai-v1"
2. care_plan: 护理计划步骤列表（中文）
3. medicine_hints: 药品建议列表（中文，仅建议非处方药，注明"仅供参考，需医生审核"）
4. follow_up: 复诊建议（中文）

重要：所有建议仅供医生参考，最终决策权在医生。

仅输出 JSON，不要输出其他内容。`

func (p *bailianProvider) Recommend(ctx context.Context, input RecommendInput) (RecommendResult, error) {
	userMsg := fmt.Sprintf("诊断: %s\n症状: %s\n去向: %s",
		input.Diagnosis, strings.Join(input.Symptoms, ", "), input.Destination)

	raw, err := p.call(ctx, recommendSystemPrompt, userMsg)
	if err != nil {
		return RecommendResult{}, fmt.Errorf("AI recommend: %w", err)
	}

	var result RecommendResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return RecommendResult{}, fmt.Errorf("parse AI recommend response: %w (raw: %s)", err, raw)
	}
	return result, nil
}

// ---- InteractionCheck ----

const interactionCheckSystemPrompt = `你是一个校园医务室的 AI 药物相互作用检查系统。检查给定药品之间是否存在相互作用。

要求：
1. has_interaction: 是否存在相互作用（布尔值）
2. risk_level: "low"、"medium" 或 "high"
3. interactions: 相互作用列表，每项包含：
   - pair: 药品对（数组，两个药品名）
   - severity: 严重程度（"low"/"medium"/"high"）
   - effect: 相互作用效果（中文）
4. advice: 用药建议列表（中文）

仅输出 JSON，不要输出其他内容。`

func (p *bailianProvider) InteractionCheck(ctx context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	userMsg := fmt.Sprintf("需要检查以下药品的相互作用：%s", strings.Join(input.Medicines, ", "))

	raw, err := p.call(ctx, interactionCheckSystemPrompt, userMsg)
	if err != nil {
		return InteractionCheckResult{}, fmt.Errorf("AI interaction check: %w", err)
	}

	var result InteractionCheckResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return InteractionCheckResult{}, fmt.Errorf("parse AI interaction check response: %w (raw: %s)", err, raw)
	}
	return result, nil
}
