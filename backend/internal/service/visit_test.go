package service

import (
	"context"
	"errors"
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
