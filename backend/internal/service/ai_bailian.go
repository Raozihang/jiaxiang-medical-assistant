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
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type chatRequest struct {
	Model         string             `json:"model"`
	Messages      []chatMessage      `json:"messages"`
	Temperature   float64            `json:"temperature,omitempty"`
	EnableSearch  bool               `json:"enable_search,omitempty"`
	SearchOptions *chatSearchOptions `json:"search_options,omitempty"`
}

type chatSearchOptions struct {
	ForcedSearch bool `json:"forced_search,omitempty"`
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

func (p *bailianProvider) call(ctx context.Context, system, user string, enableSearch bool) (string, error) {
	body := chatRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
	}
	if enableSearch {
		body.EnableSearch = true
		body.SearchOptions = &chatSearchOptions{ForcedSearch: true}
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

func extractJSON(text string) string {
	s := strings.TrimSpace(text)
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

const analyzeSystemPrompt = `你是校医务室症状分析助手。
要求：
- 只输出合法 JSON，不要输出 Markdown、解释或额外文本
- risk_level 只能是 low|medium|high
- confidence 为 0..1 的数字
- matched_signals 为字符串数组
- possible_causes、suggested_actions 中所有自然语言内容必须使用简体中文
- 保持 JSON 字段名不变`

func (p *bailianProvider) Analyze(ctx context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	userMsg := fmt.Sprintf("症状：%s\n补充描述：%s\n体温：%.1fC",
		strings.Join(input.Symptoms, "，"), input.Description, input.Temperature)
	raw, err := p.call(ctx, analyzeSystemPrompt, userMsg, false)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("AI analyze: %w", err)
	}

	var result AnalyzeResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return AnalyzeResult{}, fmt.Errorf("parse AI analyze response: %w (raw: %s)", err, raw)
	}
	return result, nil
}

const triageSystemPrompt = `你是校医务室分诊助手。
要求：
- 只输出合法 JSON，不要输出 Markdown、解释或额外文本
- triage_level 只能是 routine|priority|urgent
- destination 只能是 classroom|observation|hospital
- reason、suggested_actions 中所有自然语言内容必须使用简体中文
- review_in_minutes 为整数
- 保持 JSON 字段名不变`

func (p *bailianProvider) Triage(ctx context.Context, input TriageInput) (TriageResult, error) {
	userMsg := fmt.Sprintf("症状：%s\n补充描述：%s\n体温：%.1fC",
		strings.Join(input.Symptoms, "，"), input.Description, input.Temperature)
	raw, err := p.call(ctx, triageSystemPrompt, userMsg, false)
	if err != nil {
		return TriageResult{}, fmt.Errorf("AI triage: %w", err)
	}

	var result TriageResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return TriageResult{}, fmt.Errorf("parse AI triage response: %w (raw: %s)", err, raw)
	}
	return result, nil
}

const recommendSystemPrompt = `你是校医务室用药推荐助手。
要求：
- 只输出合法 JSON，不要输出 Markdown、解释或额外文本
- 以提供的本地药品库存与 RAG 内容为第一依据，只能推荐本地库存中的药品
- 优先选择库存大于 0、未临期、警示信息可控的药品
- medicines 中必须包含 name、dosage、frequency、duration、reason、caution
- 若没有安全可用的本地药品，medicines 返回空数组，并在 contraindications 中用简体中文说明原因
- 如启用 web search，只能用于安全性复核，不能绕过本地库存限制
- 除 plan_version、risk_flags 以及已有枚举/代码字段外，所有自然语言内容必须使用简体中文
- risk_flags 必须保持稳定的英文 snake_case 代码
- 保持 JSON 字段名不变`

func (p *bailianProvider) Recommend(ctx context.Context, input RecommendInput) (RecommendResult, error) {
	userMsg := fmt.Sprintf(
		"诊断：%s\n症状：%s\n分诊等级：%s\n去向：%s\n过敏信息：%s\n风险标记：%s\n本地库存与RAG信息：\n%s",
		input.Diagnosis,
		strings.Join(input.Symptoms, "，"),
		input.TriageLevel,
		input.Destination,
		strings.Join(input.Allergies, "，"),
		strings.Join(input.RiskFlags, ", "),
		input.RAGContext,
	)
	raw, err := p.call(ctx, recommendSystemPrompt, userMsg, input.UseWebSearch)
	if err != nil {
		return RecommendResult{}, fmt.Errorf("AI recommend: %w", err)
	}

	var result RecommendResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return RecommendResult{}, fmt.Errorf("parse AI recommend response: %w (raw: %s)", err, raw)
	}
	return result, nil
}

const interactionCheckSystemPrompt = `你是校医务室用药安全审查助手。
要求：
- 只输出合法 JSON，不要输出 Markdown、解释或额外文本
- 优先使用提供的本地药品 RAG 数据；只有启用 web search 时，才可用于安全复核
- risk_level 只能是 low|medium|high
- interactions 中 effect 必须使用简体中文；severity 建议使用 low|medium|high
- advice、warnings.title、warnings.description、warnings.suggestion 必须使用简体中文
- risk_flags 必须保持稳定的英文 snake_case 代码
- 保持 JSON 字段名不变`

func (p *bailianProvider) InteractionCheck(ctx context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	userMsg := fmt.Sprintf(
		"待核查药品：%s\n学生ID：%s\n风险标记：%s\n本地药品RAG信息：\n%s",
		strings.Join(input.Medicines, "，"),
		input.StudentID,
		strings.Join(input.RiskFlags, ", "),
		input.RAGContext,
	)
	raw, err := p.call(ctx, interactionCheckSystemPrompt, userMsg, input.UseWebSearch)
	if err != nil {
		return InteractionCheckResult{}, fmt.Errorf("AI interaction check: %w", err)
	}

	var result InteractionCheckResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &result); err != nil {
		return InteractionCheckResult{}, fmt.Errorf("parse AI interaction check response: %w (raw: %s)", err, raw)
	}
	return result, nil
}
