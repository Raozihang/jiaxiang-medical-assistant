package service

import (
	"context"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestAIAnalysisServiceQueuesAndPersistsSnapshot(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260088",
		Symptoms:    []string{"fever", "cough"},
		Description: "student reports fever after class",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	analysisService := NewAIAnalysisService(visitRepo, NewAIService())
	queued, err := analysisService.Enqueue(ctx, visit.ID, true)
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if queued.AIAnalysis.Status != "queued" || queued.AIAnalysis.QueuedAt == nil {
		t.Fatalf("expected queued analysis with timestamp, got %+v", queued.AIAnalysis)
	}

	analysisService.process(ctx, visit.ID)
	detail, err := visitRepo.GetByID(ctx, visit.ID)
	if err != nil {
		t.Fatalf("get visit failed: %v", err)
	}
	if detail.AIAnalysis.Status != "completed" {
		t.Fatalf("expected completed analysis, got %s", detail.AIAnalysis.Status)
	}
	if detail.AIAnalysis.QueuedAt == nil || detail.AIAnalysis.ProcessedAt == nil {
		t.Fatalf("expected queued_at and processed_at to be retained, got %+v", detail.AIAnalysis)
	}
	if len(detail.AIAnalysis.Analyze) == 0 || len(detail.AIAnalysis.Triage) == 0 || len(detail.AIAnalysis.Recommend) == 0 {
		t.Fatalf("expected persisted AI result payloads, got %+v", detail.AIAnalysis)
	}
}

func TestAIAnalysisServiceSendsMeasuredTemperatureToAI(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	measured := 38.0
	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:         "20260089",
		Symptoms:          []string{"fever"},
		Description:       "student reports fever",
		TemperatureStatus: "measured",
		TemperatureValue:  &measured,
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	provider := &customAIProviderStub{
		analyzeResult: AnalyzeResult{RiskLevel: "medium", Confidence: 0.8},
		recommendResult: RecommendResult{
			PlanVersion: "test",
		},
	}
	analysisService := NewAIAnalysisService(visitRepo, NewAIServiceWithProvider(provider))

	analysisService.process(ctx, visit.ID)

	if provider.lastAnalyzeInput.Temperature != measured {
		t.Fatalf("expected analyze temperature %.1f, got %.1f", measured, provider.lastAnalyzeInput.Temperature)
	}
	if provider.lastTriageInput.Temperature != measured {
		t.Fatalf("expected triage temperature %.1f, got %.1f", measured, provider.lastTriageInput.Temperature)
	}
	if provider.lastRecommendInput.Temperature != measured {
		t.Fatalf("expected recommend temperature %.1f, got %.1f", measured, provider.lastRecommendInput.Temperature)
	}
}

func TestAIAnalysisServiceUsesNormalTemperatureFallback(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:         "20260090",
		Symptoms:          []string{"cough"},
		Description:       "no fever complaint",
		TemperatureStatus: "normal",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	provider := &customAIProviderStub{
		analyzeResult: AnalyzeResult{RiskLevel: "low", Confidence: 0.7},
		recommendResult: RecommendResult{
			PlanVersion: "test",
		},
	}
	analysisService := NewAIAnalysisService(visitRepo, NewAIServiceWithProvider(provider))

	analysisService.process(ctx, visit.ID)

	if provider.lastAnalyzeInput.Temperature != 36.5 {
		t.Fatalf("expected normal fallback temperature 36.5, got %.1f", provider.lastAnalyzeInput.Temperature)
	}
	if provider.lastTriageInput.Temperature != 36.5 {
		t.Fatalf("expected normal fallback triage temperature 36.5, got %.1f", provider.lastTriageInput.Temperature)
	}
}
