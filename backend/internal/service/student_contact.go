package service

import (
	"context"
	"strings"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type StudentContactService struct {
	repo repository.StudentContactRepository
}

type StudentContactListInput struct {
	PageParams repository.PageParams
	Keyword    string
}

type UpdateStudentContactInput struct {
	StudentName      *string
	GuardianName     *string
	GuardianPhone    *string
	GuardianRelation *string
}

func NewStudentContactService(repo repository.StudentContactRepository) *StudentContactService {
	return &StudentContactService{repo: repo}
}

func (s *StudentContactService) List(ctx context.Context, input StudentContactListInput) (repository.PageResult[repository.StudentContact], error) {
	return s.repo.List(ctx, repository.StudentContactListParams{
		PageParams: input.PageParams,
		Keyword:    strings.TrimSpace(input.Keyword),
	})
}

func (s *StudentContactService) UpdateByStudentID(ctx context.Context, studentID string, input UpdateStudentContactInput) (repository.StudentContact, error) {
	trimmedStudentID := strings.TrimSpace(studentID)
	if trimmedStudentID == "" {
		return repository.StudentContact{}, ErrInvalidInput
	}
	if input.GuardianPhone != nil {
		rawPhone := strings.TrimSpace(*input.GuardianPhone)
		if rawPhone == "" {
			input.GuardianPhone = &rawPhone
		} else {
			phone := normalizeContactPhone(rawPhone)
			if phone == "" {
				return repository.StudentContact{}, ErrInvalidInput
			}
			input.GuardianPhone = &phone
		}
	}

	return s.repo.UpdateByStudentID(ctx, trimmedStudentID, repository.UpdateStudentContactInput{
		StudentName:      trimStringPtr(input.StudentName),
		GuardianName:     trimStringPtr(input.GuardianName),
		GuardianPhone:    input.GuardianPhone,
		GuardianRelation: trimStringPtr(input.GuardianRelation),
	})
}

func normalizeContactPhone(phone string) string {
	trimmed := strings.TrimSpace(phone)
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	trimmed = strings.ReplaceAll(trimmed, "-", "")
	trimmed = strings.TrimPrefix(trimmed, "+86")
	if len(trimmed) < 8 {
		return ""
	}
	return trimmed
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}
