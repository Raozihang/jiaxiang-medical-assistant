package service

import (
	"context"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestImportServiceSubmitVisitsTracksSuccessAndFailures(t *testing.T) {
	visitRepo := repository.NewMockVisitRepository()
	taskRepo := repository.NewMemoryImportTaskRepository()
	svc := NewImportService(visitRepo, taskRepo)

	createdAt := time.Now().UTC().Add(-24 * time.Hour)
	task, err := svc.SubmitVisits(context.Background(), []VisitImportItem{
		{
			StudentID:   "20269999",
			Symptoms:    []string{"fever"},
			Description: "history",
			Destination: "observation",
			CreatedAt:   &createdAt,
		},
		{
			StudentID: "   ",
		},
	})
	if err != nil {
		t.Fatalf("submit visits failed: %v", err)
	}

	if task.Total != 2 {
		t.Fatalf("expected total=2, got %d", task.Total)
	}
	if task.Success != 1 || task.Failed != 1 {
		t.Fatalf("expected success=1 failed=1, got success=%d failed=%d", task.Success, task.Failed)
	}
	if task.Status != "completed_with_errors" {
		t.Fatalf("expected completed_with_errors status, got %s", task.Status)
	}

	stored, err := svc.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if stored.ID != task.ID {
		t.Fatalf("unexpected task id: %s", stored.ID)
	}
}
