package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type SafetyService struct {
	visitRepo  repository.VisitRepository
	alertState repository.SafetyAlertStateRepository
}

type SafetyAlert struct {
	ID          string     `json:"id"`
	Rule        string     `json:"rule"`
	Status      string     `json:"status"`
	Message     string     `json:"message"`
	VisitID     string     `json:"visit_id"`
	StudentID   string     `json:"student_id"`
	CreatedAt   time.Time  `json:"created_at"`
	TriggeredAt time.Time  `json:"triggered_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

func NewSafetyService(visitRepo repository.VisitRepository, alertState repository.SafetyAlertStateRepository) *SafetyService {
	return &SafetyService{visitRepo: visitRepo, alertState: alertState}
}

func (s *SafetyService) ListAlerts(ctx context.Context, statusFilter string) ([]SafetyAlert, error) {
	visits, err := s.listAllVisits(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	filter := strings.ToLower(strings.TrimSpace(statusFilter))
	alerts := make([]SafetyAlert, 0)
	for _, visit := range visits {
		if strings.ToLower(strings.TrimSpace(visit.Destination)) != "observation" {
			continue
		}
		if visit.CreatedAt.After(now.Add(-2 * time.Hour)) {
			continue
		}

		alertID := buildObservationTimeoutAlertID(visit.ID)
		alert := SafetyAlert{
			ID:          alertID,
			Rule:        "observation_timeout",
			Status:      "open",
			Message:     "observation duration exceeded 2 hours",
			VisitID:     visit.ID,
			StudentID:   visit.StudentID,
			CreatedAt:   visit.CreatedAt,
			TriggeredAt: visit.CreatedAt.Add(2 * time.Hour),
			UpdatedAt:   now,
		}

		state, stateErr := s.alertState.GetByID(ctx, alertID)
		if stateErr == nil {
			alert.Status = state.Status
			alert.UpdatedAt = state.UpdatedAt
			alert.ResolvedAt = state.ResolvedAt
		}

		if filter != "" && filter != "all" && strings.ToLower(alert.Status) != filter {
			continue
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (s *SafetyService) ResolveAlert(ctx context.Context, id string) (SafetyAlert, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return SafetyAlert{}, ErrInvalidInput
	}

	alerts, err := s.ListAlerts(ctx, "all")
	if err != nil {
		return SafetyAlert{}, err
	}

	var target *SafetyAlert
	for idx := range alerts {
		if alerts[idx].ID == trimmedID {
			target = &alerts[idx]
			break
		}
	}
	if target == nil {
		return SafetyAlert{}, repository.ErrNotFound
	}

	now := time.Now().UTC()
	state := repository.SafetyAlertState{
		ID:         trimmedID,
		Status:     "resolved",
		ResolvedAt: &now,
		UpdatedAt:  now,
	}
	if _, err := s.alertState.Upsert(ctx, state); err != nil {
		return SafetyAlert{}, err
	}

	target.Status = state.Status
	target.ResolvedAt = state.ResolvedAt
	target.UpdatedAt = state.UpdatedAt

	return *target, nil
}

func (s *SafetyService) listAllVisits(ctx context.Context) ([]repository.Visit, error) {
	const pageSize = 200

	items := make([]repository.Visit, 0)
	page := 1
	for {
		result, err := s.visitRepo.List(ctx, repository.VisitListParams{
			PageParams: repository.PageParams{Page: page, PageSize: pageSize},
		})
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if len(items) >= int(result.Total) || len(result.Items) == 0 {
			break
		}
		page++
	}

	return items, nil
}

func buildObservationTimeoutAlertID(visitID string) string {
	return fmt.Sprintf("observation-timeout:%s", visitID)
}
