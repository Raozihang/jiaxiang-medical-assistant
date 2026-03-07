package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type MemoryImportTaskRepository struct {
	mu    sync.RWMutex
	items map[string]ImportTask
}

func NewMemoryImportTaskRepository() *MemoryImportTaskRepository {
	return &MemoryImportTaskRepository{items: map[string]ImportTask{}}
}

func (r *MemoryImportTaskRepository) Create(_ context.Context, task ImportTask) (ImportTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[task.ID] = task
	return task, nil
}

func (r *MemoryImportTaskRepository) Update(_ context.Context, task ImportTask) (ImportTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[task.ID]; !ok {
		return ImportTask{}, ErrNotFound
	}
	r.items[task.ID] = task
	return task, nil
}

func (r *MemoryImportTaskRepository) GetByID(_ context.Context, id string) (ImportTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[strings.TrimSpace(id)]
	if !ok {
		return ImportTask{}, ErrNotFound
	}
	return item, nil
}

func (r *MemoryImportTaskRepository) List(_ context.Context) ([]ImportTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]ImportTask, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items, nil
}
