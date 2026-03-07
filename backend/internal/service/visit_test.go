package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestVisitServiceUpdateTracksOutboundCall(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	callRepo := repository.NewMemoryOutboundCallRepository()
	callService := NewOutboundCallService(callRepo, visitRepo, contactRepo, NewMockOutboundCallProvider(), "external_medical_followup")
	visitService := NewVisitService(visitRepo, callService)

	visit := seedVisit(t, ctx, visitRepo)
	destination := "leave_school"

	updated, err := visitService.Update(ctx, visit.ID, UpdateVisitInput{Destination: &destination})
	if err != nil {
		t.Fatalf("update visit failed: %v", err)
	}
	if updated.Destination != destination {
		t.Fatalf("expected destination %q, got %q", destination, updated.Destination)
	}

	calls, err := callRepo.List(ctx, repository.OutboundCallListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list outbound calls failed: %v", err)
	}
	if len(calls.Items) != 1 {
		t.Fatalf("expected 1 outbound call, got %d", len(calls.Items))
	}
	if calls.Items[0].Status != "requested" {
		t.Fatalf("expected requested status, got %q", calls.Items[0].Status)
	}
}

func TestVisitServiceUpdateCreatesFailedCallWhenPhoneMissing(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	callRepo := repository.NewMemoryOutboundCallRepository()
	contactService := NewStudentContactService(contactRepo)
	callService := NewOutboundCallService(callRepo, visitRepo, contactRepo, NewMockOutboundCallProvider(), "external_medical_followup")
	visitService := NewVisitService(visitRepo, callService)

	visit := seedVisit(t, ctx, visitRepo)
	empty := ""
	if _, err := contactService.UpdateByStudentID(ctx, visit.StudentID, UpdateStudentContactInput{GuardianPhone: &empty}); err != nil {
		t.Fatalf("clear guardian phone failed: %v", err)
	}

	destination := "leave_school"
	if _, err := visitService.Update(ctx, visit.ID, UpdateVisitInput{Destination: &destination}); err != nil {
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
		t.Fatalf("expected failed status, got %q", calls.Items[0].Status)
	}
}

func TestOutboundCallRetryUsesLatestContactAndVisitDestination(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	callRepo := repository.NewMemoryOutboundCallRepository()
	contactService := NewStudentContactService(contactRepo)
	callService := NewOutboundCallService(callRepo, visitRepo, contactRepo, NewMockOutboundCallProvider(), "external_medical_followup")
	visitService := NewVisitService(visitRepo, callService)

	visit := seedVisit(t, ctx, visitRepo)
	empty := ""
	if _, err := contactService.UpdateByStudentID(ctx, visit.StudentID, UpdateStudentContactInput{GuardianPhone: &empty}); err != nil {
		t.Fatalf("clear guardian phone failed: %v", err)
	}

	firstDestination := "referred"
	if _, err := visitService.Update(ctx, visit.ID, UpdateVisitInput{Destination: &firstDestination}); err != nil {
		t.Fatalf("update visit failed: %v", err)
	}

	calls, err := callRepo.List(ctx, repository.OutboundCallListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list outbound calls failed: %v", err)
	}
	if len(calls.Items) != 1 {
		t.Fatalf("expected 1 outbound call before retry, got %d", len(calls.Items))
	}

	latestPhone := "13912345678"
	if _, err := contactService.UpdateByStudentID(ctx, visit.StudentID, UpdateStudentContactInput{GuardianPhone: &latestPhone}); err != nil {
		t.Fatalf("update guardian phone failed: %v", err)
	}

	retried, err := callService.Retry(ctx, calls.Items[0].ID)
	if err != nil {
		t.Fatalf("retry outbound call failed: %v", err)
	}
	if retried.GuardianPhone != latestPhone {
		t.Fatalf("expected latest phone %q, got %q", latestPhone, retried.GuardianPhone)
	}
	if retried.TriggerSource != "manual" {
		t.Fatalf("expected manual trigger source, got %q", retried.TriggerSource)
	}
	if retried.Status != "requested" {
		t.Fatalf("expected requested status, got %q", retried.Status)
	}
	if retried.TemplateParams == "" || !strings.Contains(retried.TemplateParams, "referred") {
		t.Fatalf("expected template params to preserve destination, got %q", retried.TemplateParams)
	}
}

func TestVisitServiceUpdateParsesAndClearsFollowUp(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	visitService := NewVisitService(visitRepo)
	visit := seedVisit(t, ctx, visitRepo)

	followUpAt := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Second)
	followUpAtText := followUpAt.Format(time.RFC3339)
	note := "bring parent reply"

	updated, err := visitService.Update(ctx, visit.ID, UpdateVisitInput{FollowUpAt: &followUpAtText, FollowUpNote: &note})
	if err != nil {
		t.Fatalf("set follow-up failed: %v", err)
	}
	if updated.FollowUpAt == nil || !updated.FollowUpAt.Equal(followUpAt) {
		t.Fatalf("expected follow_up_at %s, got %#v", followUpAt, updated.FollowUpAt)
	}
	if updated.FollowUpNote == nil || *updated.FollowUpNote != note {
		t.Fatalf("expected follow_up_note %q, got %#v", note, updated.FollowUpNote)
	}

	empty := ""
	cleared, err := visitService.Update(ctx, visit.ID, UpdateVisitInput{FollowUpAt: &empty, FollowUpNote: &empty})
	if err != nil {
		t.Fatalf("clear follow-up failed: %v", err)
	}
	if cleared.FollowUpAt != nil {
		t.Fatalf("expected follow_up_at to be cleared, got %#v", cleared.FollowUpAt)
	}
	if cleared.FollowUpNote == nil || *cleared.FollowUpNote != "" {
		t.Fatalf("expected follow_up_note to be empty string, got %#v", cleared.FollowUpNote)
	}
}

func TestVisitServiceUpdateRejectsInvalidFollowUpDatetime(t *testing.T) {
	ctx := context.Background()
	visitRepo := repository.NewMockVisitRepository()
	visitService := NewVisitService(visitRepo)
	visit := seedVisit(t, ctx, visitRepo)
	invalid := "2026-99-99 25:61"

	_, err := visitService.Update(ctx, visit.ID, UpdateVisitInput{FollowUpAt: &invalid})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func seedVisit(t *testing.T, ctx context.Context, repo repository.VisitRepository) repository.Visit {
	t.Helper()
	if err := repo.EnsureSeedData(ctx); err != nil {
		t.Fatalf("seed visits failed: %v", err)
	}
	result, err := repo.List(ctx, repository.VisitListParams{PageParams: repository.PageParams{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list visits failed: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected seeded visit")
	}
	return result.Items[0]
}
