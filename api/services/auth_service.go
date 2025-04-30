package services

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"time"

	"igaku/dtos"
	"igaku/errors"
	"igaku/models"
	"igaku/repositories"
	"igaku/utils"
)

const tokenDuration = time.Hour * 1 // TODO: Store in `.env`

type AuthService interface {
	Login(creds dtos.LoginCredentials) (string, error)
	Register(fields dtos.RegistrationFields) (string, error)
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

func (s *authService) Register(fields dtos.RegistrationFields) (string, error) {
	// TODO: Consider creating a UserRepository's method designed
	// specifically to check if a user with the given Username exists,
	// without fetching the user.
	_, err := s.repo.FindByUsername(fields.Username)

	if err == nil {
		return "", &errors.UsernameAlreadyTakenError{}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fields.Password), 2)
	usr := models.User {
		ID: uuid.New(),
		Username: fields.Username,
		Password: string(hashedPassword),
		Role: models.Patient,
	}
	err = s.repo.Persist(&usr)

	expirationTime := time.Now().Add(tokenDuration)
	return utils.GenerateJWTToken(&usr, time.Now(), expirationTime)
}
