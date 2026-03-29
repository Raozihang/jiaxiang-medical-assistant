package model

import "time"

type NotificationLog struct {
	ID        string    `gorm:"size:64;primaryKey" json:"id"`
	Channel   string    `gorm:"size:32;not null;index" json:"channel"`
	Receiver  string    `gorm:"size:128;not null;index" json:"receiver"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Status    string    `gorm:"size:32;not null;index" json:"status"`
	Error     string    `gorm:"type:text" json:"error"`
	SentAt    time.Time `gorm:"not null;index" json:"sent_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
