package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/gorm"
)

type GormOutboundCallRepository struct {
	db *gorm.DB
}

func NewGormOutboundCallRepository(db *gorm.DB) *GormOutboundCallRepository {
	return &GormOutboundCallRepository{db: db}
}

func (r *GormOutboundCallRepository) Create(ctx context.Context, input CreateOutboundCallInput) (OutboundCall, error) {
	now := time.Now().UTC()
	row := model.OutboundCall{
		ID:               uuid.NewString(),
		VisitID:          strings.TrimSpace(input.VisitID),
		StudentID:        strings.TrimSpace(input.StudentID),
		StudentName:      strings.TrimSpace(input.StudentName),
		GuardianName:     strings.TrimSpace(input.GuardianName),
		GuardianPhone:    strings.TrimSpace(input.GuardianPhone),
		GuardianRelation: strings.TrimSpace(input.GuardianRelation),
		Scenario:         strings.TrimSpace(input.Scenario),
		Provider:         strings.TrimSpace(input.Provider),
		TriggerSource:    strings.TrimSpace(input.TriggerSource),
		Status:           strings.TrimSpace(input.Status),
		Message:          strings.TrimSpace(input.Message),
		TemplateCode:     strings.TrimSpace(input.TemplateCode),
		TemplateParams:   strings.TrimSpace(input.TemplateParams),
		RequestID:        strings.TrimSpace(input.RequestID),
		CallID:           strings.TrimSpace(input.CallID),
		Error:            strings.TrimSpace(input.Error),
		ResponseRaw:      strings.TrimSpace(input.ResponseRaw),
		RetryOfID:        input.RetryOfID,
		RequestedAt:      input.RequestedAt,
		CompletedAt:      input.CompletedAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if row.RequestedAt.IsZero() {
		row.RequestedAt = now
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return OutboundCall{}, err
	}
	return toOutboundCallDTO(row), nil
}

func (r *GormOutboundCallRepository) GetByID(ctx context.Context, id string) (OutboundCall, error) {
	var row model.OutboundCall
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return OutboundCall{}, ErrNotFound
		}
		return OutboundCall{}, err
	}
	return toOutboundCallDTO(row), nil
}

func (r *GormOutboundCallRepository) List(ctx context.Context, params OutboundCallListParams) (PageResult[OutboundCall], error) {
	query := r.db.WithContext(ctx).Model(&model.OutboundCall{})
	if status := strings.TrimSpace(params.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if studentID := strings.TrimSpace(params.StudentID); studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}
	if keyword := strings.TrimSpace(params.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("student_id ILIKE ? OR student_name ILIKE ? OR guardian_phone ILIKE ? OR guardian_name ILIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[OutboundCall]{}, err
	}

	var rows []model.OutboundCall
	if err := query.Order("requested_at desc").Order("created_at desc").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&rows).Error; err != nil {
		return PageResult[OutboundCall]{}, err
	}

	items := make([]OutboundCall, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOutboundCallDTO(row))
	}

	return PageResult[OutboundCall]{Items: items, Page: params.Page, PageSize: params.PageSize, Total: total}, nil
}

func (r *GormOutboundCallRepository) FindLatestByVisitAndScenario(ctx context.Context, visitID string, scenario string) (OutboundCall, error) {
	var row model.OutboundCall
	if err := r.db.WithContext(ctx).
		Where("visit_id = ? AND scenario = ?", strings.TrimSpace(visitID), strings.TrimSpace(scenario)).
		Order("requested_at desc").
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return OutboundCall{}, ErrNotFound
		}
		return OutboundCall{}, err
	}
	return toOutboundCallDTO(row), nil
}

func (r *GormOutboundCallRepository) FindByRequestID(ctx context.Context, requestID string) (OutboundCall, error) {
	var row model.OutboundCall
	if err := r.db.WithContext(ctx).Where("request_id = ?", strings.TrimSpace(requestID)).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return OutboundCall{}, ErrNotFound
		}
		return OutboundCall{}, err
	}
	return toOutboundCallDTO(row), nil
}

func (r *GormOutboundCallRepository) UpdateStatus(ctx context.Context, id string, input UpdateOutboundCallStatusInput) (OutboundCall, error) {
	var row model.OutboundCall
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return OutboundCall{}, ErrNotFound
		}
		return OutboundCall{}, err
	}

	row.Status = strings.TrimSpace(input.Status)
	if input.RequestID != nil {
		row.RequestID = strings.TrimSpace(*input.RequestID)
	}
	if input.CallID != nil {
		row.CallID = strings.TrimSpace(*input.CallID)
	}
	if input.Error != nil {
		row.Error = strings.TrimSpace(*input.Error)
	}
	if input.ResponseRaw != nil {
		row.ResponseRaw = strings.TrimSpace(*input.ResponseRaw)
	}
	if input.CompletedAt != nil {
		completedAt := input.CompletedAt.UTC()
		row.CompletedAt = &completedAt
	}
	row.UpdatedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return OutboundCall{}, err
	}

	return toOutboundCallDTO(row), nil
}

func toOutboundCallDTO(row model.OutboundCall) OutboundCall {
	return OutboundCall{
		ID:               row.ID,
		VisitID:          row.VisitID,
		StudentID:        row.StudentID,
		StudentName:      row.StudentName,
		GuardianName:     row.GuardianName,
		GuardianPhone:    row.GuardianPhone,
		GuardianRelation: row.GuardianRelation,
		Scenario:         row.Scenario,
		Provider:         row.Provider,
		TriggerSource:    row.TriggerSource,
		Status:           row.Status,
		Message:          row.Message,
		TemplateCode:     row.TemplateCode,
		TemplateParams:   row.TemplateParams,
		RequestID:        row.RequestID,
		CallID:           row.CallID,
		Error:            row.Error,
		ResponseRaw:      row.ResponseRaw,
		RetryOfID:        row.RetryOfID,
		RequestedAt:      row.RequestedAt,
		CompletedAt:      row.CompletedAt,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
