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
	generated := make([]SafetyAlert, 0, len(visits))
	generated = append(generated, buildObservationTimeoutAlerts(visits, now)...)
	generated = append(generated, buildVisitUnclosedAlerts(visits, now)...)
	generated = append(generated, buildRepeatVisit3DAlerts(visits, now)...)

	alerts := make([]SafetyAlert, 0, len(generated))
	for _, alert := range generated {
		state, stateErr := s.alertState.GetByID(ctx, alert.ID)
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

func buildVisitUnclosedAlertID(visitID string) string {
	return fmt.Sprintf("visit-unclosed:%s", visitID)
}

func buildRepeatVisit3DAlertID(studentID string, visitID string) string {
	return fmt.Sprintf("repeat-visit-3d:%s:%s", studentID, visitID)
}

func buildObservationTimeoutAlerts(visits []repository.Visit, now time.Time) []SafetyAlert {
	alerts := make([]SafetyAlert, 0)
	for _, visit := range visits {
		if strings.ToLower(strings.TrimSpace(visit.Destination)) != "observation" {
			continue
		}
		if visit.CreatedAt.After(now.Add(-2 * time.Hour)) {
			continue
		}

		alerts = append(alerts, SafetyAlert{
			ID:          buildObservationTimeoutAlertID(visit.ID),
			Rule:        "observation_timeout",
			Status:      "open",
			Message:     "留观时间已超过2小时",
			VisitID:     visit.ID,
			StudentID:   visit.StudentID,
			CreatedAt:   visit.CreatedAt,
			TriggeredAt: visit.CreatedAt.Add(2 * time.Hour),
			UpdatedAt:   now,
		})
	}

	return alerts
}

func buildVisitUnclosedAlerts(visits []repository.Visit, now time.Time) []SafetyAlert {
	alerts := make([]SafetyAlert, 0)
	for _, visit := range visits {
		destination := strings.ToLower(strings.TrimSpace(visit.Destination))
		if destination != "" && destination != "unknown" {
			continue
		}
		if visit.CreatedAt.After(now.Add(-30 * time.Minute)) {
			continue
		}

		alerts = append(alerts, SafetyAlert{
			ID:          buildVisitUnclosedAlertID(visit.ID),
			Rule:        "visit_unclosed",
			Status:      "open",
			Message:     "就诊记录超过30分钟未关闭",
			VisitID:     visit.ID,
			StudentID:   visit.StudentID,
			CreatedAt:   visit.CreatedAt,
			TriggeredAt: visit.CreatedAt.Add(30 * time.Minute),
			UpdatedAt:   now,
		})
	}

	return alerts
}

func buildRepeatVisit3DAlerts(visits []repository.Visit, now time.Time) []SafetyAlert {
	studentVisits := make(map[string][]repository.Visit)
	studentOrder := make([]string, 0)

	for _, visit := range visits {
		studentID := strings.TrimSpace(visit.StudentID)
		if studentID == "" {
			continue
		}
		if _, exists := studentVisits[studentID]; !exists {
			studentOrder = append(studentOrder, studentID)
		}
		studentVisits[studentID] = append(studentVisits[studentID], visit)
	}

	alerts := make([]SafetyAlert, 0)
	for _, studentID := range studentOrder {
		visitsByStudent := studentVisits[studentID]
		if len(visitsByStudent) < 2 {
			continue
		}

		latest := visitsByStudent[0]
		for idx := 1; idx < len(visitsByStudent); idx++ {
			candidate := visitsByStudent[idx]
			if candidate.CreatedAt.After(latest.CreatedAt) {
				latest = candidate
				continue
			}
			if candidate.CreatedAt.Equal(latest.CreatedAt) && candidate.ID > latest.ID {
				latest = candidate
			}
		}

		windowStart := latest.CreatedAt.Add(-72 * time.Hour)
		recentCount := 0
		for _, visit := range visitsByStudent {
			if visit.CreatedAt.Before(windowStart) || visit.CreatedAt.After(latest.CreatedAt) {
				continue
			}
			recentCount++
		}
		if recentCount < 2 {
			continue
		}

		alerts = append(alerts, SafetyAlert{
			ID:          buildRepeatVisit3DAlertID(studentID, latest.ID),
			Rule:        "repeat_visit_3d",
			Status:      "open",
			Message:     fmt.Sprintf("该学生3天内已就诊%d次", recentCount),
			VisitID:     latest.ID,
			StudentID:   studentID,
			CreatedAt:   latest.CreatedAt,
			TriggeredAt: latest.CreatedAt,
			UpdatedAt:   now,
		})
	}

	return alerts
}
