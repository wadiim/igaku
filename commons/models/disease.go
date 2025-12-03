package models

import (
	"github.com/google/uuid"
)

type Disease struct {
	ID		uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	RxNormID	string	`gorm:"uniqueIndex;not null;check:rx_norm_id <> ''" json:"rx_norm_id" binding:"required" example:"D008177"`
	Name	string	`gorm:"not null;check:name <> ''" json:"name" binding:"required" example:"Lupus Vulgaris"`
}
