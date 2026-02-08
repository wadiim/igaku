package models

import (
	"github.com/google/uuid"
)

type Substance struct {
	ID     uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	RXCUI  string     `gorm:"column:rxcui;uniqueIndex;not null;check:rxcui <> ''" json:"rxcui" binding:"required" example:"282415"`
	Name   string     `gorm:"not null;check:name <> ''" json:"name" binding:"required" example:"amantadine"`
	TTY    string     `gorm:"not null;check:tty <> ''" json:"tty" binding:"required" example:"PIN"`
}
