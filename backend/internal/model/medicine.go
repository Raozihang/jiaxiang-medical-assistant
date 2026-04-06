package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Medicine struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name          string         `gorm:"size:128;not null;index" json:"name"`
	Specification string         `gorm:"size:128;not null" json:"specification"`
	Stock         int            `gorm:"not null;default:0" json:"stock"`
	SafeStock     int            `gorm:"not null;default:20" json:"safe_stock"`
	ExpiryDate    time.Time      `gorm:"type:date;not null" json:"expiry_date"`
	Warnings      datatypes.JSON `gorm:"type:jsonb" json:"warnings"`
	RecommendedDosage    string         `gorm:"size:128;not null;default:''" json:"recommended_dosage"`
	RecommendedFrequency string         `gorm:"size:128;not null;default:''" json:"recommended_frequency"`
	RecommendedDuration  string         `gorm:"size:128;not null;default:''" json:"recommended_duration"`
	UsageInstructions    string         `gorm:"size:255;not null;default:''" json:"usage_instructions"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (m *Medicine) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	return nil
}
