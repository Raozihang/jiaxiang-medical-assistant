package model

import "time"

type SafetyAlertState struct {
	ID         string     `gorm:"size:128;primaryKey" json:"id"`
	Status     string     `gorm:"size:32;not null;index" json:"status"`
	ResolvedAt *time.Time `json:"resolved_at"`
	UpdatedAt  time.Time  `gorm:"not null" json:"updated_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
