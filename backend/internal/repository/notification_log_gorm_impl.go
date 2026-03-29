package repository

import (
	"context"

	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/gorm"
)

type GormNotificationLogRepository struct {
	db *gorm.DB
}

func NewGormNotificationLogRepository(db *gorm.DB) *GormNotificationLogRepository {
	return &GormNotificationLogRepository{db: db}
}

func (r *GormNotificationLogRepository) Append(ctx context.Context, log NotificationLog) error {
	row := model.NotificationLog{
		ID:       log.ID,
		Channel:  log.Channel,
		Receiver: log.Receiver,
		Message:  log.Message,
		Status:   log.Status,
		Error:    log.Error,
		SentAt:   log.SentAt,
	}

	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormNotificationLogRepository) List(ctx context.Context) ([]NotificationLog, error) {
	var rows []model.NotificationLog
	if err := r.db.WithContext(ctx).Order("sent_at desc").Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]NotificationLog, 0, len(rows))
	for _, row := range rows {
		result = append(result, NotificationLog{
			ID:       row.ID,
			Channel:  row.Channel,
			Receiver: row.Receiver,
			Message:  row.Message,
			Status:   row.Status,
			Error:    row.Error,
			SentAt:   row.SentAt,
		})
	}

	return result, nil
}
