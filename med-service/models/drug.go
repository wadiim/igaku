package models

import (
	"github.com/google/uuid"
)

type Drug struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	RXCUI      string     `gorm:"column:rxcui;uniqueIndex;not null;check:rxcui <> ''" json:"rxcui" binding:"required" example:"1115700"`
	Name       string     `gorm:"not null;check:name <> ''" json:"name" binding:"required" example:"Lupus Vulgaris"`
	Substance  string     `gorm:"not null;check:substance <> ''" json:"substance" binding:"required" example:"hydrocodone"`
}

type DrugOrderableField string

const (
	DrugID         DrugOrderableField = "id"
	DrugName       DrugOrderableField = "name"
	SubstanceName  DrugOrderableField = "substance"
)

var DrugOrderableFieldsMap = map[string]DrugOrderableField{
	"id":        DrugID,
	"name":      DrugName,
	"substance": SubstanceName,
}
