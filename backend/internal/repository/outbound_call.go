package repository

import (
	"context"
	"time"
)

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
