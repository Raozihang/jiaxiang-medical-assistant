package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type AIAnalysisQueue interface {
	Enqueue(ctx context.Context, visitID string, clearResults bool) (repository.Visit, error)
}

type AIAnalysisService struct {
	repo      repository.VisitRepository
	aiService *AIService
	jobs      chan string
}

func NewAIAnalysisService(repo repository.VisitRepository, aiService *AIService) *AIAnalysisService {
	return &AIAnalysisService{
		repo:      repo,
		aiService: aiService,
		jobs:      make(chan string, 100),
	}
}

func (s *AIAnalysisService) Start(ctx context.Context, workers int) {
	if workers <= 0 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go s.worker(ctx)
	}
}

func (s *AIAnalysisService) Enqueue(ctx context.Context, visitID string, clearResults bool) (repository.Visit, error) {
	visitID = strings.TrimSpace(visitID)
	if visitID == "" {
		return repository.Visit{}, ErrInvalidInput
	}

	now := time.Now().UTC()
	visit, err := s.repo.UpdateAIAnalysis(ctx, visitID, repository.UpdateAIAnalysisInput{
		Status:       "queued",
		QueuedAt:     &now,
		ClearResults: clearResults,
	})
	if err != nil {
		return repository.Visit{}, err
	}

	select {
	case s.jobs <- visitID:
	default:
		go func() {
			select {
			case s.jobs <- visitID:
			case <-time.After(5 * time.Second):
				log.Printf("AI analysis queue is full, dropped visit_id=%s", visitID)
			}
		}()
	}

	return visit, nil
}

func (s *AIAnalysisService) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case visitID := <-s.jobs:
			s.process(ctx, visitID)
		}
	}
}

func (s *AIAnalysisService) process(parent context.Context, visitID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 2*time.Minute)
	defer cancel()

	if _, err := s.repo.UpdateAIAnalysis(ctx, visitID, repository.UpdateAIAnalysisInput{Status: "processing"}); err != nil {
		log.Printf("mark AI analysis processing failed visit_id=%s: %v", visitID, err)
		return
	}

	visit, err := s.repo.GetByID(ctx, visitID)
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}

	temperature := visitTemperatureForAI(visit)
	analyze, err := s.aiService.Analyze(ctx, AnalyzeInput{
		Symptoms:    visit.Symptoms,
		Description: visit.Description,
		Temperature: temperature,
	})
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}

	triage, err := s.aiService.Triage(ctx, TriageInput{
		Symptoms:    visit.Symptoms,
		Description: visit.Description,
		Temperature: temperature,
	})
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}

	diagnosis := strings.TrimSpace(visit.Diagnosis)
	if diagnosis == "" && len(analyze.PossibleCauses) > 0 {
		diagnosis = analyze.PossibleCauses[0]
	}
	recommend, err := s.aiService.Recommend(ctx, RecommendInput{
		Diagnosis:   diagnosis,
		Symptoms:    visit.Symptoms,
		Destination: triage.Destination,
		TriageLevel: triage.TriageLevel,
		Temperature: temperature,
	})
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}

	var interaction InteractionCheckResult
	medicineNames := medicineNames(recommend.Medicines)
	if len(medicineNames) > 0 {
		interaction, err = s.aiService.InteractionCheck(ctx, InteractionCheckInput{
			Medicines: medicineNames,
			StudentID: visit.StudentID,
		})
		if err != nil {
			s.markFailed(ctx, visitID, err)
			return
		}
	}

	analyzeJSON, err := json.Marshal(analyze)
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}
	triageJSON, err := json.Marshal(triage)
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}
	recommendJSON, err := json.Marshal(recommend)
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}
	interactionJSON, err := json.Marshal(interaction)
	if err != nil {
		s.markFailed(ctx, visitID, err)
		return
	}

	now := time.Now().UTC()
	if _, err := s.repo.UpdateAIAnalysis(ctx, visitID, repository.UpdateAIAnalysisInput{
		Status:      "completed",
		Analyze:     analyzeJSON,
		Triage:      triageJSON,
		Recommend:   recommendJSON,
		Interaction: interactionJSON,
		ProcessedAt: &now,
	}); err != nil {
		log.Printf("save AI analysis failed visit_id=%s: %v", visitID, err)
	}
}

func visitTemperatureForAI(visit repository.Visit) float64 {
	if visit.TemperatureValue != nil && *visit.TemperatureValue > 0 {
		return *visit.TemperatureValue
	}
	return 36.5
}

func (s *AIAnalysisService) markFailed(ctx context.Context, visitID string, err error) {
	now := time.Now().UTC()
	if _, updateErr := s.repo.UpdateAIAnalysis(ctx, visitID, repository.UpdateAIAnalysisInput{
		Status:      "failed",
		Error:       err.Error(),
		ProcessedAt: &now,
	}); updateErr != nil {
		log.Printf("mark AI analysis failed failed visit_id=%s: %v", visitID, updateErr)
	}
}
