package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestBailianServer(t *testing.T, wantResponse any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		respJSON, _ := json.Marshal(wantResponse)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: string(respJSON)}},
			},
		})
	}))
}

func TestBailianProviderAnalyze(t *testing.T) {
	fakeResult := AnalyzeResult{
		RiskLevel:        "medium",
		Confidence:       0.82,
		MatchedSignals:   []string{"发烧", "头痛"},
		PossibleCauses:   []string{"上呼吸道感染"},
		SuggestedActions: []string{"留观"},
	}
	srv := newTestBailianServer(t, fakeResult)
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.Analyze(context.Background(), AnalyzeInput{
		Symptoms:    []string{"headache", "fever"},
		Description: "student has fever",
		Temperature: 38.5,
	})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if result.RiskLevel != "medium" {
		t.Fatalf("expected medium risk, got %s", result.RiskLevel)
	}
	if len(result.PossibleCauses) == 0 {
		t.Fatal("expected possible causes")
	}
}

func TestBailianProviderTriage(t *testing.T) {
	fakeResult := TriageResult{
		TriageLevel:      "urgent",
		Destination:      "hospital",
		Reason:           "高热伴呼吸困难",
		ReviewInMinutes:  10,
		SuggestedActions: []string{"立即联系家长", "准备转院"},
	}
	srv := newTestBailianServer(t, fakeResult)
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.Triage(context.Background(), TriageInput{
		Symptoms:    []string{"difficulty breathing"},
		Description: "chest pain",
		Temperature: 39.8,
	})
	if err != nil {
		t.Fatalf("triage failed: %v", err)
	}
	if result.TriageLevel != "urgent" {
		t.Fatalf("expected urgent, got %s", result.TriageLevel)
	}
	if result.Destination != "hospital" {
		t.Fatalf("expected hospital, got %s", result.Destination)
	}
}

func TestBailianProviderRecommend(t *testing.T) {
	fakeResult := RecommendResult{
		PlanVersion:   "ai-v1",
		CarePlan:      []string{"记录症状", "补充水分"},
		MedicineHints: []string{"布洛芬混悬液（仅供参考）"},
		FollowUp:      "2小时后复查体温",
	}
	srv := newTestBailianServer(t, fakeResult)
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.Recommend(context.Background(), RecommendInput{
		Diagnosis: "上呼吸道感染",
		Symptoms:  []string{"fever", "sore throat"},
	})
	if err != nil {
		t.Fatalf("recommend failed: %v", err)
	}
	if result.PlanVersion != "ai-v1" {
		t.Fatalf("expected ai-v1, got %s", result.PlanVersion)
	}
}

func TestBailianProviderInteractionCheck(t *testing.T) {
	fakeResult := InteractionCheckResult{
		HasInteraction: true,
		RiskLevel:      "medium",
		Interactions: []MedicationInteraction{
			{Pair: []string{"布洛芬", "阿司匹林"}, Severity: "medium", Effect: "增加胃肠道出血风险"},
		},
		Advice: []string{"避免同时使用"},
	}
	srv := newTestBailianServer(t, fakeResult)
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.InteractionCheck(context.Background(), InteractionCheckInput{
		Medicines: []string{"布洛芬", "阿司匹林"},
	})
	if err != nil {
		t.Fatalf("interaction check failed: %v", err)
	}
	if !result.HasInteraction {
		t.Fatal("expected interaction")
	}
}

func TestBailianProviderHandlesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"service unavailable","code":"500"}}`))
	}))
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	_, err := provider.Analyze(context.Background(), AnalyzeInput{
		Symptoms: []string{"headache"},
	})
	if err == nil {
		t.Fatal("expected error on API failure")
	}
}

func TestBailianProviderHandlesMarkdownFence(t *testing.T) {
	fakeResult := AnalyzeResult{
		RiskLevel:  "low",
		Confidence: 0.65,
	}
	respJSON, _ := json.Marshal(fakeResult)
	wrappedContent := "```json\n" + string(respJSON) + "\n```"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: wrappedContent}},
			},
		})
	}))
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.Analyze(context.Background(), AnalyzeInput{
		Symptoms: []string{"headache"},
	})
	if err != nil {
		t.Fatalf("expected successful parse of fenced JSON: %v", err)
	}
	if result.RiskLevel != "low" {
		t.Fatalf("expected low, got %s", result.RiskLevel)
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", `{"a":1}`, `{"a":1}`},
		{"fenced", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"fenced no lang", "```\n{\"a\":1}\n```", `{"a":1}`},
		{"whitespace", "  {\"a\":1}  ", `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.want {
				t.Fatalf("extractJSON(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
