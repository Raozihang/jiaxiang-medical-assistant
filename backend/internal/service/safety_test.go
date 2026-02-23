package service

import (
	"context"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestSafetyServiceObservationTimeoutAndResolve(t *testing.T) {
	visitRepo := repository.NewMockVisitRepository()
	stateRepo := repository.NewMemorySafetyAlertStateRepository()
	svc := NewSafetyService(visitRepo, stateRepo)

	old := time.Now().UTC().Add(-3 * time.Hour)
	visit, err := visitRepo.Create(context.Background(), repository.CreateVisitInput{
		StudentID:   "20268888",
		Symptoms:    []string{"headache"},
		Description: "needs observation",
		CreatedAt:   &old,
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	alerts, err := svc.ListAlerts(context.Background(), "")
	if err != nil {
		t.Fatalf("list alerts failed: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatalf("expected at least one alert")
	}

	alertID := ""
	for _, alert := range alerts {
		if alert.VisitID == visit.ID {
			alertID = alert.ID
			break
		}
	}
	if alertID == "" {
		t.Fatalf("expected alert for imported visit")
	}

	resolved, err := svc.ResolveAlert(context.Background(), alertID)
	if err != nil {
		t.Fatalf("resolve alert failed: %v", err)
	}
	if resolved.Status != "resolved" {
		t.Fatalf("expected resolved status, got %s", resolved.Status)
	}

	alerts, err = svc.ListAlerts(context.Background(), "")
	if err != nil {
		t.Fatalf("list alerts after resolve failed: %v", err)
	}

	for _, alert := range alerts {
		if alert.ID == alertID && alert.Status != "resolved" {
			t.Fatalf("expected resolved status after update, got %s", alert.Status)
		}
	}
}
