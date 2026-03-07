package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormImportTaskRepository struct {
	db *gorm.DB
}

func NewGormImportTaskRepository(db *gorm.DB) *GormImportTaskRepository {
	return &GormImportTaskRepository{db: db}
}

func (r *GormImportTaskRepository) Create(ctx context.Context, task ImportTask) (ImportTask, error) {
	row, err := toImportTaskModel(task)
	if err != nil {
		return ImportTask{}, err
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return ImportTask{}, err
	}
	return toImportTaskDTO(row)
}

func (r *GormImportTaskRepository) Update(ctx context.Context, task ImportTask) (ImportTask, error) {
	row, err := toImportTaskModel(task)
	if err != nil {
		return ImportTask{}, err
	}
	result := r.db.WithContext(ctx).Model(&model.ImportTask{}).Where("id = ?", row.ID).Updates(map[string]any{
		"status":     row.Status,
		"total":      row.Total,
		"success":    row.Success,
		"failed":     row.Failed,
		"errors":     row.Errors,
		"updated_at": row.UpdatedAt,
	})
	if result.Error != nil {
		return ImportTask{}, result.Error
	}
	if result.RowsAffected == 0 {
		return ImportTask{}, ErrNotFound
	}
	return task, nil
}

func (r *GormImportTaskRepository) GetByID(ctx context.Context, id string) (ImportTask, error) {
	var row model.ImportTask
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ImportTask{}, ErrNotFound
		}
		return ImportTask{}, err
	}
	return toImportTaskDTO(row)
}

func (r *GormImportTaskRepository) List(ctx context.Context) ([]ImportTask, error) {
	var rows []model.ImportTask
	if err := r.db.WithContext(ctx).Order("updated_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]ImportTask, 0, len(rows))
	for _, row := range rows {
		item, err := toImportTaskDTO(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func toImportTaskModel(task ImportTask) (model.ImportTask, error) {
	errorsRaw, err := json.Marshal(task.Errors)
	if err != nil {
		return model.ImportTask{}, err
	}
	return model.ImportTask{
		ID:        task.ID,
		Status:    task.Status,
		Total:     task.Total,
		Success:   task.Success,
		Failed:    task.Failed,
		Errors:    datatypes.JSON(errorsRaw),
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

func toImportTaskDTO(row model.ImportTask) (ImportTask, error) {
	var taskErrors []ImportTaskError
	if len(row.Errors) > 0 {
		if err := json.Unmarshal(row.Errors, &taskErrors); err != nil {
			return ImportTask{}, err
		}
	}
	return ImportTask{ID: row.ID, Status: row.Status, Total: row.Total, Success: row.Success, Failed: row.Failed, Errors: taskErrors, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}
