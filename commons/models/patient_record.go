package models

import (
	"github.com/google/uuid"
)

type PatientRecord struct {
	ID		uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	NationalID	string	`gorm:"uniqueIndex;check:length(national_id)=11 AND national_id~'^[0-9]+$'" json:"national_id,omitempty" binding:"omitempty" example:"44051401458"`
}
