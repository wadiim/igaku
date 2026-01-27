package models

import (
	"github.com/google/uuid"

	"time"

	commonsModels "igaku/commons/models"
)

type MedicalHistoryItem struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	PatientID  uuid.UUID  `gorm:"type:uuid;not null"`
	DoctorID   uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`

	Patient    commonsModels.PatientRecord  `gorm:"foreignKey:PatientID"`
	Doctor     Doctor                       `gorm:"foreignKey:DoctorID"`
}
