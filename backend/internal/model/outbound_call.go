package model

import "time"

type OutboundCall struct {
	ID               string     `gorm:"size:64;primaryKey" json:"id"`
	VisitID          string     `gorm:"size:64;not null;index" json:"visit_id"`
	StudentID        string     `gorm:"size:32;not null;index" json:"student_id"`
	StudentName      string     `gorm:"size:64;not null" json:"student_name"`
	GuardianName     string     `gorm:"size:64" json:"guardian_name"`
	GuardianPhone    string     `gorm:"size:32;not null;index" json:"guardian_phone"`
	GuardianRelation string     `gorm:"size:32" json:"guardian_relation"`
	Scenario         string     `gorm:"size:64;not null;index" json:"scenario"`
	Provider         string     `gorm:"size:32;not null;index" json:"provider"`
	TriggerSource    string     `gorm:"size:32;not null" json:"trigger_source"`
	Status           string     `gorm:"size:32;not null;index" json:"status"`
	Message          string     `gorm:"type:text;not null" json:"message"`
	TemplateCode     string     `gorm:"size:128" json:"template_code"`
	TemplateParams   string     `gorm:"type:text" json:"template_params"`
	RequestID        string     `gorm:"size:128;index" json:"request_id"`
	CallID           string     `gorm:"size:128;index" json:"call_id"`
	Error            string     `gorm:"type:text" json:"error"`
	ResponseRaw      string     `gorm:"type:text" json:"response_raw"`
	RetryOfID        *string    `gorm:"size:64;index" json:"retry_of_id"`
	RequestedAt      time.Time  `gorm:"not null;index" json:"requested_at"`
	CompletedAt      *time.Time `gorm:"index" json:"completed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
