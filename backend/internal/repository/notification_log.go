package repository

import (
	"context"
	"time"
)

type NotificationLog struct {
	ID       string    `json:"id"`
	Channel  string    `json:"channel"`
	Receiver string    `json:"receiver"`
	Message  string    `json:"message"`
	Status   string    `json:"status"`
	Error    string    `json:"error,omitempty"`
	SentAt   time.Time `json:"sent_at"`
}

type NotificationLogRepository interface {
	Append(ctx context.Context, log NotificationLog) error
	List(ctx context.Context) ([]NotificationLog, error)
}
