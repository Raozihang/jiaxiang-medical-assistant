package service

import (
	"context"
	"errors"
	"testing"
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

	if result.TriageLevel != "urgent" {
		t.Fatalf("expected urgent triage, got %s", result.TriageLevel)
	}
	if result.Destination != "hospital" {
		t.Fatalf("expected hospital destination, got %s", result.Destination)
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
	if !result.HasInteraction {
		t.Fatalf("expected known interaction")
	}
	if len(result.Interactions) == 0 {
		t.Fatalf("expected at least one interaction")
	}
}

type customAIProviderStub struct {
	analyzeCalled          bool
	triageCalled           bool
	recommendCalled        bool
	interactionCheckCalled bool
	analyzeResult          AnalyzeResult
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

func (s *customAIProviderStub) Analyze(_ context.Context, _ AnalyzeInput) (AnalyzeResult, error) {
	s.analyzeCalled = true
	return s.analyzeResult, nil
}

func (s *customAIProviderStub) Triage(_ context.Context, _ TriageInput) (TriageResult, error) {
	s.triageCalled = true
	return TriageResult{}, nil
}

func (s *customAIProviderStub) Recommend(_ context.Context, _ RecommendInput) (RecommendResult, error) {
	s.recommendCalled = true
	return RecommendResult{}, nil
}

func (s *customAIProviderStub) InteractionCheck(_ context.Context, _ InteractionCheckInput) (InteractionCheckResult, error) {
	s.interactionCheckCalled = true
	return InteractionCheckResult{}, nil
}

func TestAIServiceAnalyzeDelegatesToCustomProvider(t *testing.T) {
	provider := &customAIProviderStub{
		analyzeResult: AnalyzeResult{
			RiskLevel:  "custom",
			Confidence: 0.99,
		},
	}
	svc := NewAIServiceWithProvider(provider)

	result, err := svc.Analyze(context.Background(), AnalyzeInput{
		Symptoms: []string{"headache"},
	})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	if !provider.analyzeCalled {
		t.Fatalf("expected custom provider analyze to be called")
	}
	if provider.triageCalled || provider.recommendCalled || provider.interactionCheckCalled {
		t.Fatalf("expected only analyze delegation")
	}
	if result.RiskLevel != "custom" {
		t.Fatalf("expected custom risk level, got %s", result.RiskLevel)
	}
	if result.Confidence != 0.99 {
		t.Fatalf("expected custom confidence, got %f", result.Confidence)
	}
}

func TestAIServiceFallsBackToRuleProviderWhenCustomProviderFails(t *testing.T) {
	svc := NewAIServiceWithProvider(&failingAIProviderStub{})

	analyzeResult, err := svc.Analyze(context.Background(), AnalyzeInput{
		Symptoms:    []string{"headache"},
		Description: "student feels dizzy",
		Temperature: 37.4,
	})
	if err != nil {
		t.Fatalf("expected analyze fallback to succeed, got error: %v", err)
	}
	if analyzeResult.RiskLevel == "" {
		t.Fatalf("expected analyze fallback result")
	}

	triageResult, err := svc.Triage(context.Background(), TriageInput{
		Symptoms:    []string{"headache"},
		Description: "student feels dizzy",
		Temperature: 37.4,
	})
	if err != nil {
		t.Fatalf("expected triage fallback to succeed, got error: %v", err)
	}
	if triageResult.TriageLevel == "" || triageResult.Destination == "" {
		t.Fatalf("expected triage fallback result")
	}

	recommendResult, err := svc.Recommend(context.Background(), RecommendInput{
		Diagnosis:   "upper respiratory infection",
		Symptoms:    []string{"cough"},
		Destination: "observation",
	})
	if err != nil {
		t.Fatalf("expected recommend fallback to succeed, got error: %v", err)
	}
	if recommendResult.PlanVersion != "mock-v1" {
		t.Fatalf("expected recommend fallback plan version mock-v1, got %s", recommendResult.PlanVersion)
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
