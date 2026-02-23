package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type NotificationService struct {
	logRepo repository.NotificationLogRepository
}

type SendNotificationInput struct {
	Channel  string `json:"channel"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func NewNotificationService(logRepo repository.NotificationLogRepository) *NotificationService {
	return &NotificationService{logRepo: logRepo}
}

func (s *NotificationService) Send(ctx context.Context, input SendNotificationInput) (repository.NotificationLog, error) {
	channel := strings.ToLower(strings.TrimSpace(input.Channel))
	receiver := strings.TrimSpace(input.Receiver)
	message := strings.TrimSpace(input.Message)
	now := time.Now().UTC()

	log := repository.NotificationLog{
		ID:       uuid.NewString(),
		Channel:  channel,
		Receiver: receiver,
		Message:  message,
		Status:   "failed",
		SentAt:   now,
	}

	if channel != "wechat" && channel != "dingtalk" {
		log.Error = "unsupported channel"
		_ = s.logRepo.Append(ctx, log)
		return repository.NotificationLog{}, ErrInvalidInput
	}
	if receiver == "" || message == "" {
		log.Error = "receiver and message are required"
		_ = s.logRepo.Append(ctx, log)
		return repository.NotificationLog{}, ErrInvalidInput
	}

	log.Status = "sent"

	if err := s.logRepo.Append(ctx, log); err != nil {
		return repository.NotificationLog{}, err
	}

	return log, nil
}

func (s *NotificationService) ListLogs(ctx context.Context) ([]repository.NotificationLog, error) {
	logs, err := s.logRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].SentAt.After(logs[j].SentAt)
	})

	return logs, nil
}
