package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestBailianServer(t *testing.T, wantResponse any, inspect func(chatRequest)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		if inspect != nil {
			inspect(req)
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

func TestBailianProviderRecommendUsesSearchWhenRequested(t *testing.T) {
	fakeResult := RecommendResult{
		PlanVersion: "ai-rag-v1",
		Medicines: []MedicineRecommendation{
			{Name: "Ibuprofen Tablets", Dosage: "0.2g"},
		},
	}
	srv := newTestBailianServer(t, fakeResult, func(req chatRequest) {
		if !req.EnableSearch {
			t.Fatalf("expected enable_search=true")
		}
		if req.SearchOptions == nil || !req.SearchOptions.ForcedSearch {
			t.Fatalf("expected forced search to be enabled")
		}
	})
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.Recommend(context.Background(), RecommendInput{
		Diagnosis:    "upper respiratory infection",
		Symptoms:     []string{"fever"},
		Destination:  "observation",
		UseWebSearch: true,
	})
	if err != nil {
		t.Fatalf("recommend failed: %v", err)
	}
	if len(result.Medicines) != 1 || result.Medicines[0].Name != "Ibuprofen Tablets" {
		t.Fatalf("unexpected recommend result: %+v", result)
	}
}

func TestBailianProviderInteractionCheckPlainRequest(t *testing.T) {
	fakeResult := InteractionCheckResult{
		HasInteraction: true,
		RiskLevel:      "medium",
		Interactions: []MedicationInteraction{
			{Pair: []string{"aspirin", "ibuprofen"}, Severity: "medium", Effect: "胃肠道不良反应风险增加"},
		},
	}
	srv := newTestBailianServer(t, fakeResult, func(req chatRequest) {
		if req.EnableSearch {
			t.Fatalf("did not expect search to be enabled")
		}
	})
	defer srv.Close()

	provider := NewBailianProvider("test-key", "qwen3.5-plus", srv.URL)
	result, err := provider.InteractionCheck(context.Background(), InteractionCheckInput{
		Medicines: []string{"aspirin", "ibuprofen"},
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
	_, err := provider.Analyze(context.Background(), AnalyzeInput{Symptoms: []string{"headache"}})
	if err == nil {
		t.Fatal("expected error on API failure")
	}
}

func TestBailianProviderHandlesMarkdownFence(t *testing.T) {
	fakeResult := AnalyzeResult{RiskLevel: "low", Confidence: 0.65}
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
	result, err := provider.Analyze(context.Background(), AnalyzeInput{Symptoms: []string{"headache"}})
	if err != nil {
		t.Fatalf("expected successful parse of fenced JSON: %v", err)
	}
	if result.RiskLevel != "low" {
		t.Fatalf("expected low, got %s", result.RiskLevel)
	}
}
