package service

import (
	"context"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestReportOverviewIncludesDueFollowUps(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	medicineRepo := repository.NewMockMedicineRepository()
	svc := NewReportService(visitRepo, medicineRepo)

	pastVisit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260001",
		Symptoms:    []string{"cough"},
		Description: "past follow up",
	})
	if err != nil {
		t.Fatalf("create past visit failed: %v", err)
	}

	futureVisit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260002",
		Symptoms:    []string{"fever"},
		Description: "future follow up",
	})
	if err != nil {
		t.Fatalf("create future visit failed: %v", err)
	}

	nowDueVisit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260003",
		Symptoms:    []string{"headache"},
		Description: "now follow up",
	})
	if err != nil {
		t.Fatalf("create now visit failed: %v", err)
	}

	past := time.Now().UTC().Add(-2 * time.Hour)
	if _, err := visitRepo.Update(ctx, pastVisit.ID, repository.UpdateVisitInput{
		SetFollowUpAt: true,
		FollowUpAt:    &past,
	}); err != nil {
		t.Fatalf("update past visit failed: %v", err)
	}

	future := time.Now().UTC().Add(2 * time.Hour)
	if _, err := visitRepo.Update(ctx, futureVisit.ID, repository.UpdateVisitInput{
		SetFollowUpAt: true,
		FollowUpAt:    &future,
	}); err != nil {
		t.Fatalf("update future visit failed: %v", err)
	}

	nowDue := time.Now().UTC().Add(-1 * time.Minute)
	if _, err := visitRepo.Update(ctx, nowDueVisit.ID, repository.UpdateVisitInput{
		SetFollowUpAt: true,
		FollowUpAt:    &nowDue,
	}); err != nil {
		t.Fatalf("update now-due visit failed: %v", err)
	}

	overview, err := svc.Overview(ctx)
	if err != nil {
		t.Fatalf("overview failed: %v", err)
	}

	if overview.DueFollowUps != 2 {
		t.Fatalf("expected 2 due follow ups, got %d", overview.DueFollowUps)
	}
}
