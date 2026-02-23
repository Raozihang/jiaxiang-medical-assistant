package repository

import (
	"context"
	"time"
)

type ImportTask struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Total     int               `json:"total"`
	Success   int               `json:"success"`
	Failed    int               `json:"failed"`
	Errors    []ImportTaskError `json:"errors"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ImportTaskError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

type ImportTaskRepository interface {
	Create(ctx context.Context, task ImportTask) (ImportTask, error)
	Update(ctx context.Context, task ImportTask) (ImportTask, error)
	GetByID(ctx context.Context, id string) (ImportTask, error)
	List(ctx context.Context) ([]ImportTask, error)
}
