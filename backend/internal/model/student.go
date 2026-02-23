package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Student struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	StudentID string    `gorm:"size:32;uniqueIndex;not null" json:"student_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	ClassID   uuid.UUID `gorm:"type:uuid;not null" json:"class_id"`
	Grade     string    `gorm:"size:16;not null" json:"grade"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Student) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
