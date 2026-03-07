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
		ID:          id,
		StudentID:   "20260001",
		StudentName: "Student-20260001",
		ClassName:   "Grade 3 Class 2",
		Symptoms:    []string{"fever"},
		Description: "felt unwell after PE class",
		Destination: "observation",
		CreatedAt:   now.Add(-15 * time.Minute),
		UpdatedAt:   now.Add(-15 * time.Minute),
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
	return PageResult[Visit]{Items: items[start:end], Page: params.Page, PageSize: params.PageSize, Total: int64(len(items))}, nil
}

func (r *MockVisitRepository) Create(_ context.Context, input CreateVisitInput) (Visit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	id := uuid.NewString()
	visit := Visit{ID: id, StudentID: input.StudentID, StudentName: "Student-" + input.StudentID, ClassName: "Unknown Class", Symptoms: input.Symptoms, Description: strings.TrimSpace(input.Description), Destination: "observation", CreatedAt: now, UpdatedAt: now}
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
	if input.SetFollowUpAt {
		if input.FollowUpAt != nil {
			followUpAt := input.FollowUpAt.UTC()
			visit.FollowUpAt = &followUpAt
		} else {
			visit.FollowUpAt = nil
		}
	}
	if input.FollowUpNote != nil {
		followUpNote := strings.TrimSpace(*input.FollowUpNote)
		visit.FollowUpNote = &followUpNote
	}
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

	var count int64
	for _, visit := range r.visits {
		if visit.FollowUpAt == nil {
			continue
		}
		if !visit.FollowUpAt.After(now.UTC()) {
			count++
		}
	}

	return count, nil
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
