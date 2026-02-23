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

type VisitPeriodReport struct {
	Period                  string           `json:"period"`
	StartAt                 time.Time        `json:"start_at"`
	EndAt                   time.Time        `json:"end_at"`
	TotalVisits             int64            `json:"total_visits"`
	ObservationVisits       int64            `json:"observation_visits"`
	DestinationDistribution map[string]int64 `json:"destination_distribution"`
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

func (s *ReportService) Daily(ctx context.Context) (VisitPeriodReport, error) {
	now := time.Now().UTC()
	start := dayStart(now)
	end := start.Add(24 * time.Hour)
	return s.buildVisitPeriodReport(ctx, "daily", start, end)
}

func (s *ReportService) Weekly(ctx context.Context) (VisitPeriodReport, error) {
	now := time.Now().UTC()
	start := weekStart(now)
	end := start.AddDate(0, 0, 7)
	return s.buildVisitPeriodReport(ctx, "weekly", start, end)
}

func (s *ReportService) Monthly(ctx context.Context) (VisitPeriodReport, error) {
	now := time.Now().UTC()
	start := monthStart(now)
	end := start.AddDate(0, 1, 0)
	return s.buildVisitPeriodReport(ctx, "monthly", start, end)
}

func (s *ReportService) buildVisitPeriodReport(ctx context.Context, period string, start time.Time, end time.Time) (VisitPeriodReport, error) {
	visits, err := s.listAllVisits(ctx)
	if err != nil {
		return VisitPeriodReport{}, err
	}

	distribution := map[string]int64{}
	var total int64
	var observation int64

	for _, visit := range visits {
		if visit.CreatedAt.Before(start) || !visit.CreatedAt.Before(end) {
			continue
		}
		total++

		destination := visit.Destination
		if destination == "" {
			destination = "unknown"
		}
		distribution[destination]++
		if destination == "observation" {
			observation++
		}
	}

	return VisitPeriodReport{
		Period:                  period,
		StartAt:                 start,
		EndAt:                   end,
		TotalVisits:             total,
		ObservationVisits:       observation,
		DestinationDistribution: distribution,
	}, nil
}

func (s *ReportService) listAllVisits(ctx context.Context) ([]repository.Visit, error) {
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

func weekStart(now time.Time) time.Time {
	t := dayStart(now)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

func monthStart(now time.Time) time.Time {
	t := now.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func dayStart(now time.Time) time.Time {
	t := now.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
