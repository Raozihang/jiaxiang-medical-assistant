package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestVisitServiceUpdateRejectsInvalidFollowUpAt(t *testing.T) {
	repo := repository.NewMockVisitRepository()
	svc := NewVisitService(repo)

	invalid := "not-a-time"
	_, err := svc.Update(context.Background(), "any-id", UpdateVisitInput{
		FollowUpAt: &invalid,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestVisitServiceUpdateClearsFollowUpFieldsWithEmptyValues(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMockVisitRepository()
	svc := NewVisitService(repo)

	visit, err := repo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20261234",
		Symptoms:    []string{"cough"},
		Description: "needs follow up",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	followUpAt := time.Now().UTC().Add(24 * time.Hour)
	note := "return for recheck"
	if _, err := repo.Update(ctx, visit.ID, repository.UpdateVisitInput{
		SetFollowUpAt: true,
		FollowUpAt:    &followUpAt,
		FollowUpNote:  &note,
	}); err != nil {
		t.Fatalf("prime follow up failed: %v", err)
	}

	empty := ""
	updated, err := svc.Update(ctx, visit.ID, UpdateVisitInput{
		FollowUpAt:   &empty,
		FollowUpNote: &empty,
	})
	if err != nil {
		t.Fatalf("update visit failed: %v", err)
	}

	if updated.FollowUpAt != nil {
		t.Fatalf("expected follow_up_at to be cleared")
	}
	if updated.FollowUpNote != nil {
		t.Fatalf("expected follow_up_note to be cleared")
	}
}

func TestVisitServiceUpdateTriggersOutboundCallForLeaveSchool(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	callRepo := repository.NewMemoryOutboundCallRepository()
	callService := NewOutboundCallService(callRepo, visitRepo, contactRepo, NewMockOutboundCallProvider(), "external_medical_followup")
	svc := NewVisitService(visitRepo, callService)

	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260001",
		Symptoms:    []string{"fever"},
		Description: "need external medical follow-up",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	destination := "leave_school"
	if _, err := svc.Update(ctx, visit.ID, UpdateVisitInput{Destination: &destination}); err != nil {
		t.Fatalf("update visit failed: %v", err)
	}

	calls, err := callRepo.List(ctx, repository.OutboundCallListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list outbound calls failed: %v", err)
	}
	if len(calls.Items) != 1 {
		t.Fatalf("expected 1 outbound call, got %d", len(calls.Items))
	}
	if calls.Items[0].Status != "requested" {
		t.Fatalf("expected requested status, got %s", calls.Items[0].Status)
	}
}

func TestVisitServiceUpdateRecordsFailedOutboundCallWhenPhoneMissing(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	_, err := contactRepo.UpdateByStudentID(ctx, "20260001", repository.UpdateStudentContactInput{GuardianPhone: stringPtr("")})
	if err != nil {
		t.Fatalf("clear guardian phone failed: %v", err)
	}
	callRepo := repository.NewMemoryOutboundCallRepository()
	callService := NewOutboundCallService(callRepo, visitRepo, contactRepo, NewMockOutboundCallProvider(), "external_medical_followup")
	svc := NewVisitService(visitRepo, callService)

	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260001",
		Symptoms:    []string{"fever"},
		Description: "need external medical follow-up",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	destination := "referred"
	if _, err := svc.Update(ctx, visit.ID, UpdateVisitInput{Destination: &destination}); err != nil {
		t.Fatalf("update visit failed: %v", err)
	}

	calls, err := callRepo.List(ctx, repository.OutboundCallListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list outbound calls failed: %v", err)
	}
	if len(calls.Items) != 1 {
		t.Fatalf("expected 1 outbound call, got %d", len(calls.Items))
	}
	if calls.Items[0].Status != "failed" {
		t.Fatalf("expected failed status, got %s", calls.Items[0].Status)
	}
}

func TestVisitServiceUpdateQueuesAIAnalysisWhenTemperatureChanges(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	queue := &recordingAIAnalysisQueue{repo: visitRepo}
	svc := NewVisitService(visitRepo)
	svc.SetAIAnalysisQueue(queue)

	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:         "20260001",
		Symptoms:          []string{"fever"},
		Description:       "student reports fever",
		TemperatureStatus: "due",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	measured := 38.7
	status := "measured"
	updated, err := svc.Update(ctx, visit.ID, UpdateVisitInput{
		TemperatureStatus: &status,
		TemperatureValue:  &measured,
	})
	if err != nil {
		t.Fatalf("update visit failed: %v", err)
	}

	if queue.visitID != visit.ID {
		t.Fatalf("expected AI queue visit %s, got %s", visit.ID, queue.visitID)
	}
	if !queue.clearResults {
		t.Fatal("expected updated temperature to clear stale AI results")
	}
	if updated.AIAnalysis.Status != "queued" {
		t.Fatalf("expected updated visit to expose queued AI status, got %s", updated.AIAnalysis.Status)
	}
	if updated.TemperatureValue == nil || *updated.TemperatureValue != measured {
		t.Fatalf("expected updated temperature %.1f, got %v", measured, updated.TemperatureValue)
	}
}

func TestAIAnalysisServiceUsesTemperatureRecordedAfterVisitUpdate(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:         "20260091",
		Symptoms:          []string{"fever"},
		Description:       "student reports fever",
		TemperatureStatus: "due",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	measured := 39.1
	status := "measured"
	if _, err := visitRepo.Update(ctx, visit.ID, repository.UpdateVisitInput{
		TemperatureStatus: &status,
		TemperatureValue:  &measured,
	}); err != nil {
		t.Fatalf("update visit temperature failed: %v", err)
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

func TestOutboundCallServiceRetryUsesLatestContactAndVisitDestination(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	callRepo := repository.NewMemoryOutboundCallRepository()
	callService := NewOutboundCallService(callRepo, visitRepo, contactRepo, NewMockOutboundCallProvider(), "external_medical_followup")
	svc := NewVisitService(visitRepo, callService)

	visit, err := visitRepo.Create(ctx, repository.CreateVisitInput{
		StudentID:   "20260001",
		Symptoms:    []string{"fever"},
		Description: "need external medical follow-up",
	})
	if err != nil {
		t.Fatalf("create visit failed: %v", err)
	}

	destination := "referred"
	if _, err := svc.Update(ctx, visit.ID, UpdateVisitInput{Destination: &destination}); err != nil {
		t.Fatalf("update visit failed: %v", err)
	}

	initialCalls, err := callRepo.List(ctx, repository.OutboundCallListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list outbound calls failed: %v", err)
	}
	if len(initialCalls.Items) != 1 {
		t.Fatalf("expected 1 outbound call, got %d", len(initialCalls.Items))
	}

	newPhone := "13900000002"
	if _, err := contactRepo.UpdateByStudentID(ctx, visit.StudentID, repository.UpdateStudentContactInput{GuardianPhone: &newPhone}); err != nil {
		t.Fatalf("update guardian phone failed: %v", err)
	}

	retried, err := callService.Retry(ctx, initialCalls.Items[0].ID)
	if err != nil {
		t.Fatalf("retry outbound call failed: %v", err)
	}
	if retried.GuardianPhone != newPhone {
		t.Fatalf("expected retried guardian phone %s, got %s", newPhone, retried.GuardianPhone)
	}
	if !strings.Contains(retried.Message, "转外院") {
		t.Fatalf("expected retried message to localize referred destination, got %q", retried.Message)
	}
}

func stringPtr(value string) *string {
	return &value
}

type recordingAIAnalysisQueue struct {
	repo         repository.VisitRepository
	visitID      string
	clearResults bool
}

func (q *recordingAIAnalysisQueue) Enqueue(ctx context.Context, visitID string, clearResults bool) (repository.Visit, error) {
	q.visitID = visitID
	q.clearResults = clearResults
	return q.repo.UpdateAIAnalysis(ctx, visitID, repository.UpdateAIAnalysisInput{
		Status:       "queued",
		ClearResults: clearResults,
	})
}
