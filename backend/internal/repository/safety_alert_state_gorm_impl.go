package repository

import (
	"context"
	"errors"

	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormSafetyAlertStateRepository struct {
	db *gorm.DB
}

func NewGormSafetyAlertStateRepository(db *gorm.DB) *GormSafetyAlertStateRepository {
	return &GormSafetyAlertStateRepository{db: db}
}

func (r *GormSafetyAlertStateRepository) Upsert(ctx context.Context, state SafetyAlertState) (SafetyAlertState, error) {
	row := model.SafetyAlertState{
		ID:         state.ID,
		Status:     state.Status,
		ResolvedAt: state.ResolvedAt,
		UpdatedAt:  state.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "resolved_at", "updated_at"}),
	}).Create(&row).Error; err != nil {
		return SafetyAlertState{}, err
	}

	return SafetyAlertState{
		ID:         row.ID,
		Status:     row.Status,
		ResolvedAt: row.ResolvedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *GormSafetyAlertStateRepository) GetByID(ctx context.Context, id string) (SafetyAlertState, error) {
	var row model.SafetyAlertState
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SafetyAlertState{}, ErrNotFound
		}
		return SafetyAlertState{}, err
	}

	return SafetyAlertState{
		ID:         row.ID,
		Status:     row.Status,
		ResolvedAt: row.ResolvedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}
