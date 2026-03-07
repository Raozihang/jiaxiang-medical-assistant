package service

import (
	"context"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type stubMedicineRepository struct{}

func (s *stubMedicineRepository) List(context.Context, repository.MedicineListParams) (repository.PageResult[repository.Medicine], error) {
	return repository.PageResult[repository.Medicine]{}, nil
}
func (s *stubMedicineRepository) Inbound(context.Context, repository.StockChangeInput) (repository.Medicine, error) {
	return repository.Medicine{}, nil
}
func (s *stubMedicineRepository) Outbound(context.Context, repository.StockChangeInput) (repository.Medicine, error) {
	return repository.Medicine{}, nil
}
func (s *stubMedicineRepository) CountWarnings(context.Context, time.Time) (int64, error) {
	return 3, nil
}
func (s *stubMedicineRepository) EnsureSeedData(context.Context) error { return nil }

func TestReportServiceOverviewIncludesDueFollowUps(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	if err := visitRepo.EnsureSeedData(ctx); err != nil {
		t.Fatalf("seed visits failed: %v", err)
	}

	result, err := visitRepo.List(ctx, repository.VisitListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list visits failed: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected seeded visit")
	}

	dueAt := time.Now().UTC().Add(-time.Hour)
	note := "return for follow-up"
	if _, err := visitRepo.Update(ctx, result.Items[0].ID, repository.UpdateVisitInput{SetFollowUpAt: true, FollowUpAt: &dueAt, FollowUpNote: &note}); err != nil {
		t.Fatalf("update follow-up failed: %v", err)
	}

	reportService := NewReportService(visitRepo, &stubMedicineRepository{})
	overview, err := reportService.Overview(ctx)
	if err != nil {
		t.Fatalf("overview failed: %v", err)
	}
	if overview.DueFollowUps != 1 {
		t.Fatalf("expected due follow ups to be 1, got %d", overview.DueFollowUps)
	}
}
