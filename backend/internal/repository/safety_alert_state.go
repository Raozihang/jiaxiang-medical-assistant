package repository

import (
	"context"
	"time"
)

type SafetyAlertState struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SafetyAlertStateRepository interface {
	Upsert(ctx context.Context, state SafetyAlertState) (SafetyAlertState, error)
	GetByID(ctx context.Context, id string) (SafetyAlertState, error)
}
