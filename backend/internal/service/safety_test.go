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

func TestSafetyServiceVisitUnclosedAlert(t *testing.T) {
	visitRepo := repository.NewMockVisitRepository()
	stateRepo := repository.NewMemorySafetyAlertStateRepository()
	svc := NewSafetyService(visitRepo, stateRepo)

	createdAt := time.Now().UTC().Add(-45 * time.Minute)
	visit, err := visitRepo.Create(context.Background(), repository.CreateVisitInput{
		StudentID:   "20260011",
		Symptoms:    []string{"stomachache"},
		Description: "waiting for destination",
		CreatedAt:   &createdAt,
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	unknown := "unknown"
	if _, err := visitRepo.Update(context.Background(), visit.ID, repository.UpdateVisitInput{
		Destination: &unknown,
	}); err != nil {
		t.Fatalf("update visit destination failed: %v", err)
	}

	alerts, err := svc.ListAlerts(context.Background(), "open")
	if err != nil {
		t.Fatalf("list alerts failed: %v", err)
	}

	alert := findAlertByID(alerts, buildVisitUnclosedAlertID(visit.ID))
	if alert == nil {
		t.Fatalf("expected visit_unclosed alert for visit %s", visit.ID)
	}
	if alert.Rule != "visit_unclosed" {
		t.Fatalf("expected visit_unclosed rule, got %s", alert.Rule)
	}
}

func TestSafetyServiceRepeatVisit3DAlert(t *testing.T) {
	visitRepo := repository.NewMockVisitRepository()
	stateRepo := repository.NewMemorySafetyAlertStateRepository()
	svc := NewSafetyService(visitRepo, stateRepo)

	studentID := "20260022"
	firstCreatedAt := time.Now().UTC().Add(-90 * time.Minute)
	if _, err := visitRepo.Create(context.Background(), repository.CreateVisitInput{
		StudentID:   studentID,
		Symptoms:    []string{"headache"},
		Description: "first visit",
		CreatedAt:   &firstCreatedAt,
	}); err != nil {
		t.Fatalf("create first visit failed: %v", err)
	}

	secondCreatedAt := time.Now().UTC().Add(-30 * time.Minute)
	secondVisit, err := visitRepo.Create(context.Background(), repository.CreateVisitInput{
		StudentID:   studentID,
		Symptoms:    []string{"headache"},
		Description: "second visit",
		CreatedAt:   &secondCreatedAt,
	})
	if err != nil {
		t.Fatalf("create second visit failed: %v", err)
	}

	alertID := buildRepeatVisit3DAlertID(studentID, secondVisit.ID)
	alerts, err := svc.ListAlerts(context.Background(), "open")
	if err != nil {
		t.Fatalf("list open alerts failed: %v", err)
	}

	alert := findAlertByID(alerts, alertID)
	if alert == nil {
		t.Fatalf("expected repeat_visit_3d alert with id %s", alertID)
	}
	if alert.Rule != "repeat_visit_3d" {
		t.Fatalf("expected repeat_visit_3d rule, got %s", alert.Rule)
	}

	if _, err := svc.ResolveAlert(context.Background(), alertID); err != nil {
		t.Fatalf("resolve repeat_visit_3d alert failed: %v", err)
	}

	openAlerts, err := svc.ListAlerts(context.Background(), "open")
	if err != nil {
		t.Fatalf("list open alerts after resolve failed: %v", err)
	}
	if findAlertByID(openAlerts, alertID) != nil {
		t.Fatalf("resolved alert %s should not appear in open list", alertID)
	}

	resolvedAlerts, err := svc.ListAlerts(context.Background(), "resolved")
	if err != nil {
		t.Fatalf("list resolved alerts failed: %v", err)
	}
	resolved := findAlertByID(resolvedAlerts, alertID)
	if resolved == nil {
		t.Fatalf("expected resolved alert %s", alertID)
	}
	if resolved.Status != "resolved" {
		t.Fatalf("expected resolved status, got %s", resolved.Status)
	}
}

func findAlertByID(alerts []SafetyAlert, id string) *SafetyAlert {
	for idx := range alerts {
		if alerts[idx].ID == id {
			return &alerts[idx]
		}
	}

	return nil
}
