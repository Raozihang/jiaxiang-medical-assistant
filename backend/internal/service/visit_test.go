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
