package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MockVisitRepository struct {
	mu     sync.RWMutex
	visits map[string]Visit
}

func NewMockVisitRepository() *MockVisitRepository {
	return &MockVisitRepository{visits: map[string]Visit{}}
}

func (r *MockVisitRepository) EnsureSeedData(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.visits) > 0 {
		return nil
	}

	now := time.Now().UTC()
	id := uuid.NewString()
	r.visits[id] = Visit{
		ID:                id,
		StudentID:         "20260001",
		StudentName:       "张三",
		ClassName:         "三年级2班",
		Symptoms:          []string{"发热"},
		Description:       "上午体育课后感到不适",
		Destination:       "observation",
		TemperatureStatus: "normal",
		CreatedAt:         now.Add(-15 * time.Minute),
		UpdatedAt:         now.Add(-15 * time.Minute),
	}

	return nil
}

func (r *MockVisitRepository) List(_ context.Context, params VisitListParams) (PageResult[Visit], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Visit, 0, len(r.visits))
	for _, visit := range r.visits {
		if params.StudentID != "" && visit.StudentID != params.StudentID {
			continue
		}
		items = append(items, visit)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	start, end := pageWindow(params.Page, params.PageSize, len(items))
	return PageResult[Visit]{
		Items:    items[start:end],
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int64(len(items)),
	}, nil
}

func (r *MockVisitRepository) Create(_ context.Context, input CreateVisitInput) (Visit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if input.CreatedAt != nil {
		now = input.CreatedAt.UTC()
	}
	id := uuid.NewString()
	visit := Visit{
		ID:                id,
		StudentID:         input.StudentID,
		StudentName:       "学生-" + input.StudentID,
		ClassName:         "未知班级",
		Symptoms:          input.Symptoms,
		Description:       strings.TrimSpace(input.Description),
		Destination:       "observation",
		TemperatureStatus: normalizeTemperatureStatus(input.TemperatureStatus),
		TemperatureValue:  input.TemperatureValue,
		AIAnalysis:        AIAnalysis{Status: "not_started"},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	r.visits[id] = visit

	return visit, nil
}

func (r *MockVisitRepository) GetByID(_ context.Context, id string) (Visit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visit, ok := r.visits[id]
	if !ok {
		return Visit{}, ErrNotFound
	}

	return visit, nil
}

func (r *MockVisitRepository) Update(_ context.Context, id string, input UpdateVisitInput) (Visit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	visit, ok := r.visits[id]
	if !ok {
		return Visit{}, ErrNotFound
	}

	if input.Diagnosis != nil {
		visit.Diagnosis = strings.TrimSpace(*input.Diagnosis)
	}
	if input.Prescription != nil {
		visit.Prescription = *input.Prescription
	}
	if input.Destination != nil {
		visit.Destination = strings.TrimSpace(*input.Destination)
	}
	if input.TemperatureStatus != nil {
		visit.TemperatureStatus = strings.TrimSpace(*input.TemperatureStatus)
		if visit.TemperatureStatus == "normal" {
			visit.TemperatureValue = nil
		}
	}
	if input.TemperatureValue != nil {
		value := *input.TemperatureValue
		visit.TemperatureValue = &value
		if strings.TrimSpace(visit.TemperatureStatus) == "" {
			visit.TemperatureStatus = "measured"
		}
	}
	if input.SetFollowUpAt {
		if input.FollowUpAt == nil {
			visit.FollowUpAt = nil
		} else {
			followUpAt := input.FollowUpAt.UTC()
			visit.FollowUpAt = &followUpAt
		}
	}
	if input.FollowUpNote != nil {
		note := strings.TrimSpace(*input.FollowUpNote)
		if note == "" {
			visit.FollowUpNote = nil
		} else {
			visit.FollowUpNote = &note
		}
	}
	visit.UpdatedAt = time.Now().UTC()
	r.visits[id] = visit

	return visit, nil
}

func (r *MockVisitRepository) UpdateAIAnalysis(_ context.Context, id string, input UpdateAIAnalysisInput) (Visit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	visit, ok := r.visits[id]
	if !ok {
		return Visit{}, ErrNotFound
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = visit.AIAnalysis.Status
	}
	if status == "" {
		status = "not_started"
	}

	analysis := visit.AIAnalysis
	analysis.Status = status
	analysis.Error = strings.TrimSpace(input.Error)
	if input.QueuedAt != nil {
		analysis.QueuedAt = input.QueuedAt
	}
	if input.ProcessedAt != nil || status == "queued" {
		analysis.ProcessedAt = input.ProcessedAt
	}
	if input.ClearResults {
		analysis.Analyze = nil
		analysis.Triage = nil
		analysis.Recommend = nil
		analysis.Interaction = nil
	}
	if input.Analyze != nil {
		analysis.Analyze = append(analysis.Analyze[:0], input.Analyze...)
	}
	if input.Triage != nil {
		analysis.Triage = append(analysis.Triage[:0], input.Triage...)
	}
	if input.Recommend != nil {
		analysis.Recommend = append(analysis.Recommend[:0], input.Recommend...)
	}
	if input.Interaction != nil {
		analysis.Interaction = append(analysis.Interaction[:0], input.Interaction...)
	}

	visit.AIAnalysis = analysis
	visit.UpdatedAt = time.Now().UTC()
	r.visits[id] = visit

	return visit, nil
}

func (r *MockVisitRepository) CountToday(_ context.Context, now time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	start := dayStart(now)
	var count int64
	for _, visit := range r.visits {
		if !visit.CreatedAt.Before(start) {
			count++
		}
	}

	return count, nil
}

func (r *MockVisitRepository) CountObservationToday(_ context.Context, now time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	start := dayStart(now)
	var count int64
	for _, visit := range r.visits {
		if visit.Destination != "observation" {
			continue
		}
		if !visit.CreatedAt.Before(start) {
			count++
		}
	}

	return count, nil
}

func (r *MockVisitRepository) CountDueFollowUps(_ context.Context, now time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := now.UTC()
	var count int64
	for _, visit := range r.visits {
		if visit.FollowUpAt == nil {
			continue
		}
		if !visit.FollowUpAt.After(cutoff) {
			count++
		}
	}

	return count, nil
}

func dayStart(now time.Time) time.Time {
	ts := now.UTC()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
}

func normalizeTemperatureStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "normal"
	}
	return status
}

func pageWindow(page int, pageSize int, total int) (int, int) {
	if total == 0 {
		return 0, 0
	}

	start := (page - 1) * pageSize
	if start >= total {
		return total, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return start, end
}
