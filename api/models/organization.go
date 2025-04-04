package models

import (
	"github.com/google/uuid"
)

type Organization struct {
	ID	uuid.UUID	`gorm:type:uuid;default:uuid_generate_v4()"`
	Name	string		`json:"name"`
	Address	string		`json:"address"`
}
