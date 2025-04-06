package models

import (
	"github.com/google/uuid"
)

type Organization struct {
	ID	uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id"`
	Name	string		`json:"name"`
	Address	string		`json:"address"`
}
