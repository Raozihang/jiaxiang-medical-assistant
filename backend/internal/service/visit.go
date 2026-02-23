package service

import (
	"context"
	"strings"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type VisitService struct {
	repo repository.VisitRepository
}

type VisitListInput struct {
	PageParams repository.PageParams
	StudentID  string
}

type CreateVisitInput struct {
	StudentID   string
	Symptoms    []string
	Description string
}

type UpdateVisitInput struct {
	Diagnosis    *string
	Prescription *[]string
	Destination  *string
}

func NewVisitService(repo repository.VisitRepository) *VisitService {
	return &VisitService{repo: repo}
}

func (s *VisitService) EnsureSeedData(ctx context.Context) error {
	return s.repo.EnsureSeedData(ctx)
}

func (s *VisitService) List(ctx context.Context, input VisitListInput) (repository.PageResult[repository.Visit], error) {
	return s.repo.List(ctx, repository.VisitListParams{
		PageParams: input.PageParams,
		StudentID:  strings.TrimSpace(input.StudentID),
	})
}

func (s *VisitService) Create(ctx context.Context, input CreateVisitInput) (repository.Visit, error) {
	symptoms := make([]string, 0, len(input.Symptoms))
	for _, symptom := range input.Symptoms {
		symptom = strings.TrimSpace(symptom)
		if symptom == "" {
			continue
		}
		symptoms = append(symptoms, symptom)
	}

	return s.repo.Create(ctx, repository.CreateVisitInput{
		StudentID:   strings.TrimSpace(input.StudentID),
		Symptoms:    symptoms,
		Description: strings.TrimSpace(input.Description),
	})
}

func (s *VisitService) GetByID(ctx context.Context, id string) (repository.Visit, error) {
	return s.repo.GetByID(ctx, strings.TrimSpace(id))
}

func (s *VisitService) Update(ctx context.Context, id string, input UpdateVisitInput) (repository.Visit, error) {
	return s.repo.Update(ctx, strings.TrimSpace(id), repository.UpdateVisitInput{
		Diagnosis:    input.Diagnosis,
		Prescription: input.Prescription,
		Destination:  input.Destination,
	})
}
