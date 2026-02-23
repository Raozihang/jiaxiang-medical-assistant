package service

import (
	"context"
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
