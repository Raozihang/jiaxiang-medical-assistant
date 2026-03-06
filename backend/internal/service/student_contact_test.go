package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestStudentContactServiceAllowsClearingGuardianPhone(t *testing.T) {
	repo := repository.NewMemoryStudentContactRepository()
	svc := NewStudentContactService(repo)
	empty := ""

	updated, err := svc.UpdateByStudentID(context.Background(), "20260001", UpdateStudentContactInput{
		GuardianPhone: &empty,
	})
	if err != nil {
		t.Fatalf("update contact failed: %v", err)
	}
	if updated.GuardianPhone != "" {
		t.Fatalf("expected phone to be cleared, got %q", updated.GuardianPhone)
	}
}

func TestStudentContactServiceRejectsInvalidGuardianPhone(t *testing.T) {
	repo := repository.NewMemoryStudentContactRepository()
	svc := NewStudentContactService(repo)
	invalid := "123"

	_, err := svc.UpdateByStudentID(context.Background(), "20260001", UpdateStudentContactInput{
		GuardianPhone: &invalid,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
