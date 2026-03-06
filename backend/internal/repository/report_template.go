package repository

import (
	"context"
	"time"
)

type ReportTemplate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Period    string    `json:"period"`
	Columns   []string  `json:"columns"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReportSchedule struct {
	ID         string     `json:"id"`
	TemplateID string     `json:"template_id"`
	CronExpr   string     `json:"cron_expr"`
	Enabled    bool       `json:"enabled"`
	LastRunAt  *time.Time `json:"last_run_at"`
	NextRunAt  *time.Time `json:"next_run_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ReportTemplateRepository interface {
	Create(ctx context.Context, tpl ReportTemplate) error
	Get(ctx context.Context, id string) (ReportTemplate, error)
	List(ctx context.Context) ([]ReportTemplate, error)
	Update(ctx context.Context, tpl ReportTemplate) error
	Delete(ctx context.Context, id string) error
}

type ReportScheduleRepository interface {
	Create(ctx context.Context, sched ReportSchedule) error
	Get(ctx context.Context, id string) (ReportSchedule, error)
	List(ctx context.Context) ([]ReportSchedule, error)
	Update(ctx context.Context, sched ReportSchedule) error
	Delete(ctx context.Context, id string) error
	ListDue(ctx context.Context, now time.Time) ([]ReportSchedule, error)
}
