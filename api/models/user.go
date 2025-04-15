package models

import (
	"github.com/google/uuid"
)

type Role string

const (
	Patient	Role = "patient"
	Doctor	Role = "doctor"
	Admin	Role = "admin"
)

type User struct {
	ID		uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	Username	string		`gorm:"uniqueIndex;not null" json:"username" example:"jdoe"`
	Password	string		`gorm:"not null" json:"password" example:"$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq"`
	Role		Role		`gorm:"type:role;not null" json:"role" example:"patient"`
}
