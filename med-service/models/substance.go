package models

import (
	"github.com/google/uuid"
)

type Substance struct {
	ID	uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	RxClassID	string	`gorm:"uniqueIndex;not null;check:rx_class_id <> ''" json:"rx_class_id" binding:"required" example:"1234567"`
	Name	string `gorm:"not null;check:name <> ''" json:"name" binding:"required" example:"amantadine"`
	SubstanceType	string `gorm:"not null;check:substance_type <> ''" json:"substance_type" binding:"required" example:"PIN"`
}
