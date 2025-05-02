package utils

import (
	"github.com/golang-jwt/jwt/v5"

	"os"
	"time"

	"igaku/auth-service/errors"
	"igaku/commons/models"
)

var jwtSecretKey = []byte(os.Getenv("SECRET_KEY"))

type Claims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(user *models.User, issued time.Time, expires time.Time) (string, error) {
	claims := &Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:	user.ID.String(),
			IssuedAt:	jwt.NewNumericDate(issued),
			ExpiresAt:	jwt.NewNumericDate(expires),
			Issuer:		"igaku",
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(jwtSecretKey)
	if err != nil {
		return "", &errors.TokenGenerationError{}
	}

	return token, nil
}
