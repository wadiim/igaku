package models

import (
	"github.com/google/uuid"
)

type Setting struct {
	ID	uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id"`
	Key	string		`json:"key"`
	Value	string		`json:"value"`
}
