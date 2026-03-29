package model

import "time"

type ReportTemplate struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Period    string    `json:"period" gorm:"size:20;not null"` // daily, weekly, monthly
	Columns   string    `json:"columns" gorm:"type:text"`       // JSON array of column keys
	Title     string    `json:"title" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReportSchedule struct {
	ID         string    `json:"id" gorm:"primaryKey;size:36"`
	TemplateID string    `json:"template_id" gorm:"size:36;not null"`
	CronExpr   string    `json:"cron_expr" gorm:"size:50;not null"` // e.g. "0 18 * * *"
	Enabled    bool      `json:"enabled" gorm:"default:true"`
	LastRunAt  *time.Time `json:"last_run_at"`
	NextRunAt  *time.Time `json:"next_run_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
