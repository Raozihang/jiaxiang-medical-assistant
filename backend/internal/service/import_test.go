package service

import (
	"context"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestImportServiceSubmitVisits(t *testing.T) {
	visitRepo := repository.NewMockVisitRepository()
	taskRepo := repository.NewMemoryImportTaskRepository()
	service := NewImportService(visitRepo, taskRepo)

	task, err := service.SubmitVisits(context.Background(), []VisitImportItem{{StudentID: "20260001", Symptoms: []string{"cough"}, Description: "imported visit"}, {StudentID: "", Description: "invalid"}})
	if err != nil {
		t.Fatalf("submit visits failed: %v", err)
	}
	if task.Total != 2 || task.Success != 1 || task.Failed != 1 {
		t.Fatalf("unexpected task result: %#v", task)
	}
	if len(task.Errors) != 1 {
		t.Fatalf("expected 1 task error, got %d", len(task.Errors))
	}
}
