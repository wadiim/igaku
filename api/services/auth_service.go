package services

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"os"
	"time"

	"igaku/dto"
	"igaku/errors"
	"igaku/repositories"
	"igaku/utils"
)

var jwtSecretKey = []byte(os.Getenv("SECRET_KEY"))
const tokenDuration = time.Hour * 1 // TODO: Store in `.env`

type AuthService interface {
	Login(creds dto.LoginCredentials) (string, error)
}

type authService struct {
	repo repositories.UserRepository
}

func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Login(creds dto.LoginCredentials) (string, error) {
	usr, err := s.repo.FindByUsername(creds.Username)

	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.Password),
		[]byte(creds.Password),
	)
	if err != nil {
		return "", &errors.InvalidUsernameOrPasswordError{}
	}

	expirationTime := time.Now().Add(tokenDuration)
	claims := &utils.Claims{
		Role: usr.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:	usr.ID.String(),
			ExpiresAt:	jwt.NewNumericDate(expirationTime),
			IssuedAt:	jwt.NewNumericDate(time.Now()),
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
