package models

import (
	"github.com/google/uuid"
)

type Doctor struct {
	ID	uuid.UUID `gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	MedicalHistoryItems []MedicalHistoryItem `gorm:"foreignKey:DoctorID"`
}
