package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type VisitService struct {
	repo                repository.VisitRepository
	outboundCallService *OutboundCallService
	realtimeHub         *RealtimeHub
	aiAnalysisQueue     AIAnalysisQueue
}

type VisitListInput struct {
	PageParams repository.PageParams
	StudentID  string
}

type CreateVisitInput struct {
	StudentID         string
	Symptoms          []string
	Description       string
	TemperatureStatus string
	TemperatureValue  *float64
	CreatedAt         *time.Time
}

type UpdateVisitInput struct {
	Diagnosis         *string
	Prescription      *[]string
	Destination       *string
	TemperatureStatus *string
	TemperatureValue  *float64
	FollowUpAt        *string
	FollowUpNote      *string
}

func NewVisitService(repo repository.VisitRepository, outboundCallService ...*OutboundCallService) *VisitService {
	service := &VisitService{repo: repo}
	if len(outboundCallService) > 0 {
		service.outboundCallService = outboundCallService[0]
	}
	return service
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

	visit, err := s.repo.Create(ctx, repository.CreateVisitInput{
		StudentID:         strings.TrimSpace(input.StudentID),
		Symptoms:          symptoms,
		Description:       strings.TrimSpace(input.Description),
		TemperatureStatus: normalizeVisitTemperatureStatus(input.TemperatureStatus),
		TemperatureValue:  input.TemperatureValue,
		CreatedAt:         input.CreatedAt,
	})
	if err != nil {
		return repository.Visit{}, err
	}
	if s.aiAnalysisQueue != nil {
		if queued, queueErr := s.aiAnalysisQueue.Enqueue(context.WithoutCancel(ctx), visit.ID, true); queueErr != nil {
			log.Printf("enqueue AI analysis failed visit_id=%s: %v", visit.ID, queueErr)
		} else {
			visit = queued
		}
	}
	s.broadcastVisitsSnapshot(ctx, "created", &visit)

	return visit, nil
}

func normalizeVisitTemperatureStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "normal"
	}
	return status
}

func (s *VisitService) GetByID(ctx context.Context, id string) (repository.Visit, error) {
	return s.repo.GetByID(ctx, strings.TrimSpace(id))
}

func (s *VisitService) Update(ctx context.Context, id string, input UpdateVisitInput) (repository.Visit, error) {
	repoInput := repository.UpdateVisitInput{
		Diagnosis:         input.Diagnosis,
		Prescription:      input.Prescription,
		Destination:       input.Destination,
		TemperatureStatus: input.TemperatureStatus,
		TemperatureValue:  input.TemperatureValue,
		FollowUpNote:      input.FollowUpNote,
	}

	if input.FollowUpAt != nil {
		repoInput.SetFollowUpAt = true
		followUpAtRaw := strings.TrimSpace(*input.FollowUpAt)
		if followUpAtRaw != "" {
			followUpAt, err := time.Parse(time.RFC3339, followUpAtRaw)
			if err != nil {
				return repository.Visit{}, ErrInvalidInput
			}
			followUpAtUTC := followUpAt.UTC()
			repoInput.FollowUpAt = &followUpAtUTC
		}
	}

	visit, err := s.repo.Update(ctx, strings.TrimSpace(id), repoInput)
	if err != nil {
		return repository.Visit{}, err
	}
	if s.outboundCallService != nil {
		s.outboundCallService.TrackVisitUpdate(ctx, visit)
	}
	if s.aiAnalysisQueue != nil && visitUpdateShouldRefreshAI(input) {
		if queued, queueErr := s.aiAnalysisQueue.Enqueue(context.WithoutCancel(ctx), visit.ID, true); queueErr != nil {
			log.Printf("enqueue AI analysis after visit update failed visit_id=%s: %v", visit.ID, queueErr)
		} else {
			visit = queued
		}
	}
	s.broadcastVisitsSnapshot(ctx, "updated", &visit)

	return visit, nil
}

func visitUpdateShouldRefreshAI(input UpdateVisitInput) bool {
	return input.TemperatureStatus != nil || input.TemperatureValue != nil
}

func (s *VisitService) RegenerateAIAnalysis(ctx context.Context, id string) (repository.Visit, error) {
	if s.aiAnalysisQueue == nil {
		return repository.Visit{}, ErrInvalidInput
	}
	visit, err := s.aiAnalysisQueue.Enqueue(context.WithoutCancel(ctx), strings.TrimSpace(id), true)
	if err != nil {
		return repository.Visit{}, err
	}
	s.broadcastVisitsSnapshot(ctx, "ai_regenerate_queued", &visit)
	return visit, nil
}
