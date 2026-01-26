package dtos

import (
	"github.com/google/uuid"
)

type PatientDetails struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	Username	string `json:"username" binding:"required" example:"jdoe"`
	Email	string `json:"email" binding:"required" example:"jdoe@mail.com"`
	NationalID	string `json:"national_id" binding:"required" example:"44051401458"`
}
