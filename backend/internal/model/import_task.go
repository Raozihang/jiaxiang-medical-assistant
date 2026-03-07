package model

import (
	"time"

	"gorm.io/datatypes"
)

type ImportTask struct {
	ID        string         `gorm:"size:64;primaryKey" json:"id"`
	Status    string         `gorm:"size:32;not null;index" json:"status"`
	Total     int            `gorm:"not null;default:0" json:"total"`
	Success   int            `gorm:"not null;default:0" json:"success"`
	Failed    int            `gorm:"not null;default:0" json:"failed"`
	Errors    datatypes.JSON `gorm:"type:jsonb" json:"errors"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
