package service

import (
	"context"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type ReportService struct {
	visitRepo    repository.VisitRepository
	medicineRepo repository.MedicineRepository
}

type Overview struct {
	TodayVisits         int64 `json:"today_visits"`
	ObservationStudents int64 `json:"observation_students"`
	StockWarnings       int64 `json:"stock_warnings"`
	DueFollowUps        int64 `json:"due_follow_ups"`
}

func NewReportService(visitRepo repository.VisitRepository, medicineRepo repository.MedicineRepository) *ReportService {
	return &ReportService{visitRepo: visitRepo, medicineRepo: medicineRepo}
}

func (s *ReportService) Overview(ctx context.Context) (Overview, error) {
	now := time.Now().UTC()
	todayVisits, err := s.visitRepo.CountToday(ctx, now)
	if err != nil {
		return Overview{}, err
	}
	observation, err := s.visitRepo.CountObservationToday(ctx, now)
	if err != nil {
		return Overview{}, err
	}
	warnings, err := s.medicineRepo.CountWarnings(ctx, now)
	if err != nil {
		return Overview{}, err
	}
	dueFollowUps, err := s.visitRepo.CountDueFollowUps(ctx, now)
	if err != nil {
		return Overview{}, err
	}

	return Overview{
		TodayVisits:         todayVisits,
		ObservationStudents: observation,
		StockWarnings:       warnings,
		DueFollowUps:        dueFollowUps,
	}, nil
}
