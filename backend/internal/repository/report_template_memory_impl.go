package repository

import (
	"context"
	"sort"
	"sync"
	"time"
)

// ---- Template ----

type memoryReportTemplateRepository struct {
	mu    sync.RWMutex
	items map[string]ReportTemplate
}

func NewMemoryReportTemplateRepository() ReportTemplateRepository {
	return &memoryReportTemplateRepository{items: make(map[string]ReportTemplate)}
}

func (r *memoryReportTemplateRepository) Create(_ context.Context, tpl ReportTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[tpl.ID] = tpl
	return nil
}

func (r *memoryReportTemplateRepository) Get(_ context.Context, id string) (ReportTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tpl, ok := r.items[id]
	if !ok {
		return ReportTemplate{}, ErrNotFound
	}
	return tpl, nil
}

func (r *memoryReportTemplateRepository) List(_ context.Context) ([]ReportTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ReportTemplate, 0, len(r.items))
	for _, tpl := range r.items {
		result = append(result, tpl)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

func (r *memoryReportTemplateRepository) Update(_ context.Context, tpl ReportTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[tpl.ID]; !ok {
		return ErrNotFound
	}
	r.items[tpl.ID] = tpl
	return nil
}

func (r *memoryReportTemplateRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

// ---- Schedule ----

type memoryReportScheduleRepository struct {
	mu    sync.RWMutex
	items map[string]ReportSchedule
}

func NewMemoryReportScheduleRepository() ReportScheduleRepository {
	return &memoryReportScheduleRepository{items: make(map[string]ReportSchedule)}
}

func (r *memoryReportScheduleRepository) Create(_ context.Context, sched ReportSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[sched.ID] = sched
	return nil
}

func (r *memoryReportScheduleRepository) Get(_ context.Context, id string) (ReportSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sched, ok := r.items[id]
	if !ok {
		return ReportSchedule{}, ErrNotFound
	}
	return sched, nil
}

func (r *memoryReportScheduleRepository) List(_ context.Context) ([]ReportSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ReportSchedule, 0, len(r.items))
	for _, sched := range r.items {
		result = append(result, sched)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

func (r *memoryReportScheduleRepository) Update(_ context.Context, sched ReportSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[sched.ID]; !ok {
		return ErrNotFound
	}
	r.items[sched.ID] = sched
	return nil
}

func (r *memoryReportScheduleRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

func (r *memoryReportScheduleRepository) ListDue(_ context.Context, now time.Time) ([]ReportSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ReportSchedule, 0)
	for _, sched := range r.items {
		if sched.Enabled && sched.NextRunAt != nil && !sched.NextRunAt.After(now) {
			result = append(result, sched)
		}
	}
	return result, nil
}
