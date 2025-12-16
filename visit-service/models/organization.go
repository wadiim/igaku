package models

import (
	"github.com/google/uuid"

	"igaku/commons/dtos"
)

type Organization struct {
	ID	uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id" example:"86e6a1f3-d7aa-4e74-a20a-ea78bc13340b"`
	Name	string		`json:"name" example:"The Lowell General Hospital"`
	dtos.Location		`gorm:"embedded;embeddedPrefix:loc_" json:"location"`
}
