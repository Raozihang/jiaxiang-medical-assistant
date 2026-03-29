package repository

import (
	"context"
	"sync"
)

type MemoryNotificationLogRepository struct {
	mu   sync.RWMutex
	logs []NotificationLog
}

func NewMemoryNotificationLogRepository() *MemoryNotificationLogRepository {
	return &MemoryNotificationLogRepository{logs: []NotificationLog{}}
}

func (r *MemoryNotificationLogRepository) Append(_ context.Context, log NotificationLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = append(r.logs, log)
	return nil
}

func (r *MemoryNotificationLogRepository) List(_ context.Context) ([]NotificationLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]NotificationLog, len(r.logs))
	copy(result, r.logs)
	return result, nil
}
