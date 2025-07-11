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

type UserOrderableField string

const (
	ID		UserOrderableField = "id"
	Username	UserOrderableField = "username"
	Email		UserOrderableField = "email"
)

var UserOrderableFieldsMap = map[string]UserOrderableField{
	"id": ID,
	"username": Username,
	"email": Email,
}

type User struct {
	ID		uuid.UUID	`gorm:"type:uuid;primary_key;" json:"id" binding:"required" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	Username	string		`gorm:"uniqueIndex;not null;check:username <> ''" json:"username" binding:"required" example:"jdoe"`
	Email		string		`gorm:"not null;uniqueIndex;check:email <> ''" json:"email" binding:"required" example:"jdoe@mail.com"`
	Password	string		`gorm:"not null" json:"password" binding:"required" example:"$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq"`
	Role		Role		`gorm:"type:role;not null" json:"role" binding:"required" example:"patient"`
}
