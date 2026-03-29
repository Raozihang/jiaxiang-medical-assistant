package repository

import (
	"context"
	"sort"
	"sync"
)

type MemoryImportTaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]ImportTask
}

func NewMemoryImportTaskRepository() *MemoryImportTaskRepository {
	return &MemoryImportTaskRepository{tasks: map[string]ImportTask{}}
}

func (r *MemoryImportTaskRepository) Create(_ context.Context, task ImportTask) (ImportTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID] = task
	return task, nil
}

func (r *MemoryImportTaskRepository) Update(_ context.Context, task ImportTask) (ImportTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.ID]; !ok {
		return ImportTask{}, ErrNotFound
	}
	r.tasks[task.ID] = task
	return task, nil
}

func (r *MemoryImportTaskRepository) GetByID(_ context.Context, id string) (ImportTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return ImportTask{}, ErrNotFound
	}

	return task, nil
}

func (r *MemoryImportTaskRepository) List(_ context.Context) ([]ImportTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ImportTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		result = append(result, task)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	return result, nil
}
