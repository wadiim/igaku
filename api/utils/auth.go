package utils

import (
	"github.com/golang-jwt/jwt/v5"

	"igaku/models"
)

type Claims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}
