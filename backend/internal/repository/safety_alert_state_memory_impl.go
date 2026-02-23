package repository

import (
	"context"
	"sync"
)

type MemorySafetyAlertStateRepository struct {
	mu     sync.RWMutex
	states map[string]SafetyAlertState
}

func NewMemorySafetyAlertStateRepository() *MemorySafetyAlertStateRepository {
	return &MemorySafetyAlertStateRepository{states: map[string]SafetyAlertState{}}
}

func (r *MemorySafetyAlertStateRepository) Upsert(_ context.Context, state SafetyAlertState) (SafetyAlertState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.states[state.ID] = state
	return state, nil
}

func (r *MemorySafetyAlertStateRepository) GetByID(_ context.Context, id string) (SafetyAlertState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.states[id]
	if !ok {
		return SafetyAlertState{}, ErrNotFound
	}

	return state, nil
}
