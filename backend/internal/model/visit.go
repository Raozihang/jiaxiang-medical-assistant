package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Visit struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	StudentID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"student_id"`
	DoctorID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"doctor_id"`
	Symptoms     datatypes.JSON `gorm:"type:jsonb" json:"symptoms"`
	Description  string         `gorm:"type:text" json:"description"`
	Diagnosis    string         `gorm:"type:text" json:"diagnosis"`
	Prescription datatypes.JSON `gorm:"type:jsonb" json:"prescription"`
	Destination  string         `gorm:"size:32" json:"destination"`
	FollowUpAt   *time.Time     `gorm:"type:timestamptz" json:"follow_up_at"`
	FollowUpNote *string        `gorm:"type:text" json:"follow_up_note"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (v *Visit) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}

	return nil
}
