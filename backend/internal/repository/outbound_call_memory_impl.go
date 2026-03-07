package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryOutboundCallRepository struct {
	mu    sync.RWMutex
	items map[string]OutboundCall
}

func NewMemoryOutboundCallRepository() *MemoryOutboundCallRepository {
	return &MemoryOutboundCallRepository{items: map[string]OutboundCall{}}
}

func (r *MemoryOutboundCallRepository) Create(_ context.Context, input CreateOutboundCallInput) (OutboundCall, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	item := OutboundCall{
		ID:               uuid.NewString(),
		VisitID:          strings.TrimSpace(input.VisitID),
		StudentID:        strings.TrimSpace(input.StudentID),
		StudentName:      strings.TrimSpace(input.StudentName),
		GuardianName:     strings.TrimSpace(input.GuardianName),
		GuardianPhone:    strings.TrimSpace(input.GuardianPhone),
		GuardianRelation: strings.TrimSpace(input.GuardianRelation),
		Scenario:         strings.TrimSpace(input.Scenario),
		Provider:         strings.TrimSpace(input.Provider),
		TriggerSource:    strings.TrimSpace(input.TriggerSource),
		Status:           strings.TrimSpace(input.Status),
		Message:          strings.TrimSpace(input.Message),
		TemplateCode:     strings.TrimSpace(input.TemplateCode),
		TemplateParams:   strings.TrimSpace(input.TemplateParams),
		RequestID:        strings.TrimSpace(input.RequestID),
		CallID:           strings.TrimSpace(input.CallID),
		Error:            strings.TrimSpace(input.Error),
		ResponseRaw:      strings.TrimSpace(input.ResponseRaw),
		RetryOfID:        input.RetryOfID,
		RequestedAt:      input.RequestedAt,
		CompletedAt:      input.CompletedAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if item.RequestedAt.IsZero() {
		item.RequestedAt = now
	}
	r.items[item.ID] = item
	return item, nil
}

func (r *MemoryOutboundCallRepository) GetByID(_ context.Context, id string) (OutboundCall, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[strings.TrimSpace(id)]
	if !ok {
		return OutboundCall{}, ErrNotFound
	}
	return item, nil
}

func (r *MemoryOutboundCallRepository) List(_ context.Context, params OutboundCallListParams) (PageResult[OutboundCall], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]OutboundCall, 0, len(r.items))
	for _, item := range r.items {
		if status := strings.TrimSpace(params.Status); status != "" && item.Status != status {
			continue
		}
		if studentID := strings.TrimSpace(params.StudentID); studentID != "" && item.StudentID != studentID {
			continue
		}
		if keyword := strings.TrimSpace(params.Keyword); keyword != "" {
			target := strings.Join([]string{item.StudentID, item.StudentName, item.GuardianName, item.GuardianPhone}, " ")
			if !strings.Contains(target, keyword) {
				continue
			}
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].RequestedAt.After(items[j].RequestedAt)
	})

	start, end := pageWindow(params.Page, params.PageSize, len(items))
	return PageResult[OutboundCall]{Items: items[start:end], Page: params.Page, PageSize: params.PageSize, Total: int64(len(items))}, nil
}

func (r *MemoryOutboundCallRepository) FindLatestByVisitAndScenario(_ context.Context, visitID string, scenario string) (OutboundCall, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var (
		matched OutboundCall
		found   bool
	)
	for _, item := range r.items {
		if item.VisitID != strings.TrimSpace(visitID) || item.Scenario != strings.TrimSpace(scenario) {
			continue
		}
		if !found || item.RequestedAt.After(matched.RequestedAt) {
			matched = item
			found = true
		}
	}
	if !found {
		return OutboundCall{}, ErrNotFound
	}
	return matched, nil
}

func (r *MemoryOutboundCallRepository) FindByRequestID(_ context.Context, requestID string) (OutboundCall, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.RequestID == strings.TrimSpace(requestID) {
			return item, nil
		}
	}
	return OutboundCall{}, ErrNotFound
}

func (r *MemoryOutboundCallRepository) UpdateStatus(_ context.Context, id string, input UpdateOutboundCallStatusInput) (OutboundCall, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[strings.TrimSpace(id)]
	if !ok {
		return OutboundCall{}, ErrNotFound
	}
	item.Status = strings.TrimSpace(input.Status)
	if input.RequestID != nil {
		item.RequestID = strings.TrimSpace(*input.RequestID)
	}
	if input.CallID != nil {
		item.CallID = strings.TrimSpace(*input.CallID)
	}
	if input.Error != nil {
		item.Error = strings.TrimSpace(*input.Error)
	}
	if input.ResponseRaw != nil {
		item.ResponseRaw = strings.TrimSpace(*input.ResponseRaw)
	}
	if input.CompletedAt != nil {
		completedAt := input.CompletedAt.UTC()
		item.CompletedAt = &completedAt
	}
	item.UpdatedAt = time.Now().UTC()

	r.items[item.ID] = item
	return item, nil
}
