package repository

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/gorm"
)

type GormReportTemplateRepository struct {
	db *gorm.DB
}

func NewGormReportTemplateRepository(db *gorm.DB) *GormReportTemplateRepository {
	return &GormReportTemplateRepository{db: db}
}

func (r *GormReportTemplateRepository) Create(ctx context.Context, tpl ReportTemplate) error {
	row, err := toReportTemplateModel(tpl)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormReportTemplateRepository) Get(ctx context.Context, id string) (ReportTemplate, error) {
	var row model.ReportTemplate
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ReportTemplate{}, ErrNotFound
		}
		return ReportTemplate{}, err
	}

	return toReportTemplateDTO(row)
}

func (r *GormReportTemplateRepository) List(ctx context.Context) ([]ReportTemplate, error) {
	var rows []model.ReportTemplate
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ReportTemplate, 0, len(rows))
	for _, row := range rows {
		item, err := toReportTemplateDTO(row)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *GormReportTemplateRepository) Update(ctx context.Context, tpl ReportTemplate) error {
	row, err := toReportTemplateModel(tpl)
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Model(&model.ReportTemplate{}).Where("id = ?", row.ID).Updates(map[string]any{
		"name":       row.Name,
		"period":     row.Period,
		"columns":    row.Columns,
		"title":      row.Title,
		"updated_at": row.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *GormReportTemplateRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.ReportTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

type GormReportScheduleRepository struct {
	db *gorm.DB
}

func NewGormReportScheduleRepository(db *gorm.DB) *GormReportScheduleRepository {
	return &GormReportScheduleRepository{db: db}
}

func (r *GormReportScheduleRepository) Create(ctx context.Context, sched ReportSchedule) error {
	return r.db.WithContext(ctx).Create(&model.ReportSchedule{
		ID:         sched.ID,
		TemplateID: sched.TemplateID,
		CronExpr:   sched.CronExpr,
		Enabled:    sched.Enabled,
		LastRunAt:  sched.LastRunAt,
		NextRunAt:  sched.NextRunAt,
		CreatedAt:  sched.CreatedAt,
		UpdatedAt:  sched.UpdatedAt,
	}).Error
}

func (r *GormReportScheduleRepository) Get(ctx context.Context, id string) (ReportSchedule, error) {
	var row model.ReportSchedule
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ReportSchedule{}, ErrNotFound
		}
		return ReportSchedule{}, err
	}

	return toReportScheduleDTO(row), nil
}

func (r *GormReportScheduleRepository) List(ctx context.Context) ([]ReportSchedule, error) {
	var rows []model.ReportSchedule
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ReportSchedule, 0, len(rows))
	for _, row := range rows {
		result = append(result, toReportScheduleDTO(row))
	}

	return result, nil
}

func (r *GormReportScheduleRepository) Update(ctx context.Context, sched ReportSchedule) error {
	result := r.db.WithContext(ctx).Model(&model.ReportSchedule{}).Where("id = ?", sched.ID).Updates(map[string]any{
		"template_id": sched.TemplateID,
		"cron_expr":   sched.CronExpr,
		"enabled":     sched.Enabled,
		"last_run_at": sched.LastRunAt,
		"next_run_at": sched.NextRunAt,
		"updated_at":  sched.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *GormReportScheduleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.ReportSchedule{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *GormReportScheduleRepository) ListDue(ctx context.Context, now time.Time) ([]ReportSchedule, error) {
	var rows []model.ReportSchedule
	if err := r.db.WithContext(ctx).
		Where("enabled = ? AND next_run_at IS NOT NULL AND next_run_at <= ?", true, now).
		Order("next_run_at asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ReportSchedule, 0, len(rows))
	for _, row := range rows {
		result = append(result, toReportScheduleDTO(row))
	}

	return result, nil
}

func toReportTemplateModel(tpl ReportTemplate) (model.ReportTemplate, error) {
	columnsRaw, err := json.Marshal(tpl.Columns)
	if err != nil {
		return model.ReportTemplate{}, err
	}

	return model.ReportTemplate{
		ID:        tpl.ID,
		Name:      tpl.Name,
		Period:    tpl.Period,
		Columns:   string(columnsRaw),
		Title:     tpl.Title,
		CreatedAt: tpl.CreatedAt,
		UpdatedAt: tpl.UpdatedAt,
	}, nil
}

func toReportTemplateDTO(row model.ReportTemplate) (ReportTemplate, error) {
	var columns []string
	if row.Columns != "" {
		if err := json.Unmarshal([]byte(row.Columns), &columns); err != nil {
			return ReportTemplate{}, err
		}
	}

	return ReportTemplate{
		ID:        row.ID,
		Name:      row.Name,
		Period:    row.Period,
		Columns:   columns,
		Title:     row.Title,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func toReportScheduleDTO(row model.ReportSchedule) ReportSchedule {
	return ReportSchedule{
		ID:         row.ID,
		TemplateID: row.TemplateID,
		CronExpr:   row.CronExpr,
		Enabled:    row.Enabled,
		LastRunAt:  row.LastRunAt,
		NextRunAt:  row.NextRunAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func deleteSchedulesByTemplate(ctx context.Context, repo ReportScheduleRepository, templateID string) error {
	schedules, err := repo.List(ctx)
	if err != nil {
		return err
	}

	ids := make([]string, 0)
	for _, sched := range schedules {
		if sched.TemplateID == templateID {
			ids = append(ids, sched.ID)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := repo.Delete(ctx, id); err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
	}

	return nil
}
