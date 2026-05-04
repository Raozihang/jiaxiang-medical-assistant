package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestAIServiceTriageHighRisk(t *testing.T) {
	svc := NewAIService()
	result, err := svc.Triage(context.Background(), TriageInput{
		Symptoms:    []string{"difficulty breathing", "chest pain"},
		Description: "student reports chest pain",
		Temperature: 39.8,
	})
	if err != nil {
		t.Fatalf("triage failed: %v", err)
	}
	if result.TriageLevel != "urgent" || result.Destination != "hospital" {
		t.Fatalf("unexpected triage result: %+v", result)
	}
}

func TestAIServiceRecommendInjectsInventorySafety(t *testing.T) {
	repo := repository.NewMockMedicineRepository()
	if err := repo.EnsureSeedData(context.Background()); err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	svc := NewAIServiceWithDependencies(defaultAIProvider, repo)
	result, err := svc.Recommend(context.Background(), RecommendInput{
		Diagnosis:   "upper respiratory infection",
		Symptoms:    []string{"fever"},
		Destination: "observation",
	})
	if err != nil {
		t.Fatalf("recommend failed: %v", err)
	}
	if len(result.Medicines) == 0 {
		t.Fatalf("expected at least one medicine recommendation: %+v", result)
	}
	if result.Medicines[0].Dosage == "" || len(result.InventoryBasis) == 0 {
		t.Fatalf("expected dosage and inventory basis to be filled: %+v", result)
	}
}

func TestAIServiceInteractionCheckDetectsKnownPair(t *testing.T) {
	svc := NewAIService()
	result, err := svc.InteractionCheck(context.Background(), InteractionCheckInput{
		Medicines: []string{"aspirin", "ibuprofen"},
	})
	if err != nil {
		t.Fatalf("interaction check failed: %v", err)
	}
	if !result.HasInteraction || len(result.Interactions) == 0 {
		t.Fatalf("expected known interaction: %+v", result)
	}
}

type customAIProviderStub struct {
	analyzeCalled          bool
	triageCalled           bool
	recommendCalled        bool
	interactionCheckCalled bool
	analyzeResult          AnalyzeResult
	recommendResult        RecommendResult
	interactionResult      InteractionCheckResult
	lastAnalyzeInput       AnalyzeInput
	lastTriageInput        TriageInput
	lastRecommendInput     RecommendInput
}

type failingAIProviderStub struct{}

func (s *failingAIProviderStub) Analyze(_ context.Context, _ AnalyzeInput) (AnalyzeResult, error) {
	return AnalyzeResult{}, errors.New("provider unavailable")
}

func (s *failingAIProviderStub) Triage(_ context.Context, _ TriageInput) (TriageResult, error) {
	return TriageResult{}, errors.New("provider unavailable")
}

func (s *failingAIProviderStub) Recommend(_ context.Context, _ RecommendInput) (RecommendResult, error) {
	return RecommendResult{}, errors.New("provider unavailable")
}

func (s *failingAIProviderStub) InteractionCheck(_ context.Context, _ InteractionCheckInput) (InteractionCheckResult, error) {
	return InteractionCheckResult{}, errors.New("provider unavailable")
}

func (s *customAIProviderStub) Analyze(_ context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	s.analyzeCalled = true
	s.lastAnalyzeInput = input
	return s.analyzeResult, nil
}

func (s *customAIProviderStub) Triage(_ context.Context, input TriageInput) (TriageResult, error) {
	s.triageCalled = true
	s.lastTriageInput = input
	return TriageResult{}, nil
}

func (s *customAIProviderStub) Recommend(_ context.Context, input RecommendInput) (RecommendResult, error) {
	s.recommendCalled = true
	s.lastRecommendInput = input
	return s.recommendResult, nil
}

func (s *customAIProviderStub) InteractionCheck(_ context.Context, _ InteractionCheckInput) (InteractionCheckResult, error) {
	s.interactionCheckCalled = true
	return s.interactionResult, nil
}

func TestAIServiceAnalyzeDelegatesToCustomProvider(t *testing.T) {
	provider := &customAIProviderStub{
		analyzeResult: AnalyzeResult{RiskLevel: "custom", Confidence: 0.99},
	}
	svc := NewAIServiceWithProvider(provider)
	result, err := svc.Analyze(context.Background(), AnalyzeInput{Symptoms: []string{"headache"}})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !provider.analyzeCalled || provider.triageCalled || provider.recommendCalled || provider.interactionCheckCalled {
		t.Fatalf("expected only analyze delegation")
	}
	if result.RiskLevel != "custom" || result.Confidence != 0.99 {
		t.Fatalf("unexpected analyze result: %+v", result)
	}
}

func TestAIServiceFallsBackToRuleProviderWhenCustomProviderFails(t *testing.T) {
	repo := repository.NewMockMedicineRepository()
	if err := repo.EnsureSeedData(context.Background()); err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	svc := NewAIServiceWithDependencies(&failingAIProviderStub{}, repo)
	recommendResult, err := svc.Recommend(context.Background(), RecommendInput{
		Diagnosis:   "upper respiratory infection",
		Symptoms:    []string{"cough"},
		Destination: "observation",
	})
	if err != nil {
		t.Fatalf("expected recommend fallback to succeed, got error: %v", err)
	}
	if recommendResult.PlanVersion == "" {
		t.Fatalf("expected recommend fallback result")
	}

	interactionResult, err := svc.InteractionCheck(context.Background(), InteractionCheckInput{
		Medicines: []string{"aspirin", "ibuprofen"},
	})
	if err != nil {
		t.Fatalf("expected interaction fallback to succeed, got error: %v", err)
	}
	if !interactionResult.HasInteraction {
		t.Fatalf("expected interaction fallback to detect known pair")
	}
}
