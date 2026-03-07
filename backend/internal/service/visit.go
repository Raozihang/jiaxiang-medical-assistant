package service

import (
	"context"
	"strings"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type VisitService struct {
	repo                repository.VisitRepository
	outboundCallService *OutboundCallService
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

func NewVisitService(repo repository.VisitRepository, outboundCallService ...*OutboundCallService) *VisitService {
	var outbound *OutboundCallService
	if len(outboundCallService) > 0 {
		outbound = outboundCallService[0]
	}

	return &VisitService{repo: repo, outboundCallService: outbound}
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
	visit, err := s.repo.Update(ctx, strings.TrimSpace(id), repository.UpdateVisitInput{
		Diagnosis:    input.Diagnosis,
		Prescription: input.Prescription,
		Destination:  input.Destination,
	})
	if err != nil {
		return repository.Visit{}, err
	}

	if s.outboundCallService != nil {
		s.outboundCallService.TrackVisitUpdate(ctx, visit)
	}

	return visit, nil
}
