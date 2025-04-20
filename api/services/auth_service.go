package services

import (
	"golang.org/x/crypto/bcrypt"

	"time"

	"igaku/dtos"
	"igaku/errors"
	"igaku/repositories"
	"igaku/utils"
)

const tokenDuration = time.Hour * 1 // TODO: Store in `.env`

type AuthService interface {
	Login(creds dtos.LoginCredentials) (string, error)
}

type authService struct {
	repo repositories.UserRepository
}

func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Login(creds dtos.LoginCredentials) (string, error) {
	usr, err := s.repo.FindByUsername(creds.Username)

	if err != nil {
		return "", &errors.InvalidUsernameOrPasswordError{}
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.Password),
		[]byte(creds.Password),
	)
	if err != nil {
		return "", &errors.InvalidUsernameOrPasswordError{}
	}

	expirationTime := time.Now().Add(tokenDuration)
	return utils.GenerateJWTToken(usr, time.Now(), expirationTime)
}
