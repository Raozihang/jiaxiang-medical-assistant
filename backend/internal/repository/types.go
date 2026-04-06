package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("资源未找到")
	ErrInsufficientStock = errors.New("库存不足")
)

type PageParams struct {
	Page     int
	PageSize int
}

type PageResult[T any] struct {
	Items    []T
	Page     int
	PageSize int
	Total    int64
}

type Visit struct {
	ID           string     `json:"id"`
	StudentID    string     `json:"student_id"`
	StudentName  string     `json:"student_name"`
	ClassName    string     `json:"class_name"`
	Symptoms     []string   `json:"symptoms"`
	Description  string     `json:"description"`
	Diagnosis    string     `json:"diagnosis"`
	Prescription []string   `json:"prescription"`
	Destination  string     `json:"destination"`
	FollowUpAt   *time.Time `json:"follow_up_at"`
	FollowUpNote *string    `json:"follow_up_note"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type VisitListParams struct {
	PageParams
	StudentID string
}

type CreateVisitInput struct {
	StudentID   string
	Symptoms    []string
	Description string
	CreatedAt   *time.Time
}

type UpdateVisitInput struct {
	Diagnosis     *string
	Prescription  *[]string
	Destination   *string
	FollowUpAt    *time.Time
	SetFollowUpAt bool
	FollowUpNote  *string
}

type VisitRepository interface {
	List(ctx context.Context, params VisitListParams) (PageResult[Visit], error)
	Create(ctx context.Context, input CreateVisitInput) (Visit, error)
	GetByID(ctx context.Context, id string) (Visit, error)
	Update(ctx context.Context, id string, input UpdateVisitInput) (Visit, error)
	CountToday(ctx context.Context, now time.Time) (int64, error)
	CountObservationToday(ctx context.Context, now time.Time) (int64, error)
	CountDueFollowUps(ctx context.Context, now time.Time) (int64, error)
	EnsureSeedData(ctx context.Context) error
}

type Medicine struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Specification  string    `json:"specification"`
	Stock          int       `json:"stock"`
	SafeStock      int       `json:"safe_stock"`
	ExpiryDate     time.Time `json:"expiry_date"`
	Warnings       []string  `json:"warnings"`
	RecommendedDosage    string    `json:"recommended_dosage"`
	RecommendedFrequency string    `json:"recommended_frequency"`
	RecommendedDuration  string    `json:"recommended_duration"`
	UsageInstructions    string    `json:"usage_instructions"`
	IsLowStock     bool      `json:"is_low_stock"`
	IsExpiringSoon bool      `json:"is_expiring_soon"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MedicineListParams struct {
	PageParams
}

type StockChangeInput struct {
	MedicineID string
	Quantity   int
}

type MedicineRepository interface {
	List(ctx context.Context, params MedicineListParams) (PageResult[Medicine], error)
	ListAll(ctx context.Context) ([]Medicine, error)
	Inbound(ctx context.Context, input StockChangeInput) (Medicine, error)
	Outbound(ctx context.Context, input StockChangeInput) (Medicine, error)
	CountWarnings(ctx context.Context, now time.Time) (int64, error)
	EnsureSeedData(ctx context.Context) error
}
