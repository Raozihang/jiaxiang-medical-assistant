package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInsufficientStock = errors.New("insufficient stock")
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
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	StudentName  string    `json:"student_name"`
	ClassName    string    `json:"class_name"`
	Symptoms     []string  `json:"symptoms"`
	Description  string    `json:"description"`
	Diagnosis    string    `json:"diagnosis"`
	Prescription []string  `json:"prescription"`
	Destination  string    `json:"destination"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type VisitListParams struct {
	PageParams
	StudentID string
}

type CreateVisitInput struct {
	StudentID   string
	Symptoms    []string
	Description string
}

type UpdateVisitInput struct {
	Diagnosis    *string
	Prescription *[]string
	Destination  *string
}

type VisitRepository interface {
	List(ctx context.Context, params VisitListParams) (PageResult[Visit], error)
	Create(ctx context.Context, input CreateVisitInput) (Visit, error)
	GetByID(ctx context.Context, id string) (Visit, error)
	Update(ctx context.Context, id string, input UpdateVisitInput) (Visit, error)
	CountToday(ctx context.Context, now time.Time) (int64, error)
	CountObservationToday(ctx context.Context, now time.Time) (int64, error)
	EnsureSeedData(ctx context.Context) error
}

type StudentContact struct {
	StudentID        string `json:"student_id"`
	StudentName      string `json:"student_name"`
	GuardianName     string `json:"guardian_name"`
	GuardianPhone    string `json:"guardian_phone"`
	GuardianRelation string `json:"guardian_relation"`
}

type StudentContactListParams struct {
	PageParams
	Keyword string
}

type UpdateStudentContactInput struct {
	StudentName      *string
	GuardianName     *string
	GuardianPhone    *string
	GuardianRelation *string
}

type StudentContactRepository interface {
	List(ctx context.Context, params StudentContactListParams) (PageResult[StudentContact], error)
	GetByStudentID(ctx context.Context, studentID string) (StudentContact, error)
	UpdateByStudentID(ctx context.Context, studentID string, input UpdateStudentContactInput) (StudentContact, error)
}

type OutboundCall struct {
	ID               string     `json:"id"`
	VisitID          string     `json:"visit_id"`
	StudentID        string     `json:"student_id"`
	StudentName      string     `json:"student_name"`
	GuardianName     string     `json:"guardian_name"`
	GuardianPhone    string     `json:"guardian_phone"`
	GuardianRelation string     `json:"guardian_relation"`
	Scenario         string     `json:"scenario"`
	Provider         string     `json:"provider"`
	TriggerSource    string     `json:"trigger_source"`
	Status           string     `json:"status"`
	Message          string     `json:"message"`
	TemplateCode     string     `json:"template_code"`
	TemplateParams   string     `json:"template_params"`
	RequestID        string     `json:"request_id"`
	CallID           string     `json:"call_id"`
	Error            string     `json:"error,omitempty"`
	ResponseRaw      string     `json:"response_raw,omitempty"`
	RetryOfID        *string    `json:"retry_of_id,omitempty"`
	RequestedAt      time.Time  `json:"requested_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type OutboundCallListParams struct {
	PageParams
	Status    string
	StudentID string
	Keyword   string
}

type CreateOutboundCallInput struct {
	VisitID          string
	StudentID        string
	StudentName      string
	GuardianName     string
	GuardianPhone    string
	GuardianRelation string
	Scenario         string
	Provider         string
	TriggerSource    string
	Status           string
	Message          string
	TemplateCode     string
	TemplateParams   string
	RequestID        string
	CallID           string
	Error            string
	ResponseRaw      string
	RetryOfID        *string
	RequestedAt      time.Time
	CompletedAt      *time.Time
}

type UpdateOutboundCallStatusInput struct {
	Status      string
	RequestID   *string
	CallID      *string
	Error       *string
	ResponseRaw *string
	CompletedAt *time.Time
}

type OutboundCallRepository interface {
	Create(ctx context.Context, input CreateOutboundCallInput) (OutboundCall, error)
	GetByID(ctx context.Context, id string) (OutboundCall, error)
	List(ctx context.Context, params OutboundCallListParams) (PageResult[OutboundCall], error)
	FindLatestByVisitAndScenario(ctx context.Context, visitID string, scenario string) (OutboundCall, error)
	FindByRequestID(ctx context.Context, requestID string) (OutboundCall, error)
	UpdateStatus(ctx context.Context, id string, input UpdateOutboundCallStatusInput) (OutboundCall, error)
}

type Medicine struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Specification  string    `json:"specification"`
	Stock          int       `json:"stock"`
	SafeStock      int       `json:"safe_stock"`
	ExpiryDate     time.Time `json:"expiry_date"`
	Warnings       []string  `json:"warnings"`
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
	Inbound(ctx context.Context, input StockChangeInput) (Medicine, error)
	Outbound(ctx context.Context, input StockChangeInput) (Medicine, error)
	CountWarnings(ctx context.Context, now time.Time) (int64, error)
	EnsureSeedData(ctx context.Context) error
}
