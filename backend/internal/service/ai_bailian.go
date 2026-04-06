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

const analyzeSystemPrompt = `You are a school clinic analysis assistant.
Return JSON only with:
- risk_level: low|medium|high
- confidence: number 0..1
- matched_signals: string[]
- possible_causes: string[]
- suggested_actions: string[]`

func (p *bailianProvider) Analyze(ctx context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	userMsg := fmt.Sprintf("Symptoms: %s\nDescription: %s\nTemperature: %.1fC",
		strings.Join(input.Symptoms, ", "), input.Description, input.Temperature)
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

const triageSystemPrompt = `You are a school clinic triage assistant.
Return JSON only with:
- triage_level: routine|priority|urgent
- destination: classroom|observation|hospital
- reason: string
- review_in_minutes: integer
- suggested_actions: string[]`

func (p *bailianProvider) Triage(ctx context.Context, input TriageInput) (TriageResult, error) {
	userMsg := fmt.Sprintf("Symptoms: %s\nDescription: %s\nTemperature: %.1fC",
		strings.Join(input.Symptoms, ", "), input.Description, input.Temperature)
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

const recommendSystemPrompt = `You are a school clinic medication recommendation assistant.
Use the provided local medicine inventory as the primary RAG source.
Rules:
- recommend only from the provided local inventory
- prefer medicines with stock > 0 and not expiring soon
- include dosage, frequency, duration, reason, caution
- if no safe local option exists, return an empty medicines array and explain why
- if web search is enabled, use it only as a safety cross-check
Return JSON only with:
- plan_version: string
- medicines: [{name,dosage,frequency,duration,reason,caution}]
- advice: string[]
- contraindications: string[]
- risk_flags: string[]`

func (p *bailianProvider) Recommend(ctx context.Context, input RecommendInput) (RecommendResult, error) {
	userMsg := fmt.Sprintf(
		"Diagnosis: %s\nSymptoms: %s\nTriage: %s\nDestination: %s\nAllergies: %s\nRisk flags: %s\nLocal inventory RAG:\n%s",
		input.Diagnosis,
		strings.Join(input.Symptoms, ", "),
		input.TriageLevel,
		input.Destination,
		strings.Join(input.Allergies, ", "),
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

const interactionCheckSystemPrompt = `You are a school clinic medication safety assistant.
Use the provided local medicine RAG data first, and use web search only when enabled for safety verification.
Return JSON only with:
- has_interaction: boolean
- risk_level: low|medium|high
- interactions: [{pair,severity,effect}]
- advice: string[]
- warnings: [{title,severity,description,suggestion}]
- risk_flags: string[]`

func (p *bailianProvider) InteractionCheck(ctx context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	userMsg := fmt.Sprintf(
		"Medicines: %s\nStudent ID: %s\nRisk flags: %s\nLocal medicine RAG:\n%s",
		strings.Join(input.Medicines, ", "),
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
