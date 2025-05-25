package services

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"fmt"
	"time"

	"igaku/auth-service/clients"
	"igaku/auth-service/dtos"
	"igaku/auth-service/errors"
	"igaku/commons/models"
	"igaku/commons/utils"
)

const tokenDuration = time.Hour * 1 // TODO: Store in `.env`

type AuthService interface {
	Login(creds dtos.LoginCredentials) (string, error)
	Register(fields dtos.RegistrationFields) (string, error)
}

type authService struct {
	userClient clients.UserClient
	mailClient clients.MailClient
}

func NewAuthService(
	userClient clients.UserClient, mailClient clients.MailClient,
) AuthService {
	return &authService{userClient: userClient, mailClient: mailClient}
}

func (s *authService) Login(creds dtos.LoginCredentials) (string, error) {
	usr, err := s.userClient.FindByUsername(creds.Username)

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

// TODO: Check for email duplication.
func (s *authService) Register(fields dtos.RegistrationFields) (string, error) {
	// TODO: Consider creating a UserRepository's method designed
	// specifically to check if a user with the given Username exists,
	// without fetching the user.
	_, err := s.userClient.FindByUsername(fields.Username)

	if err == nil {
		return "", &errors.UsernameAlreadyTakenError{}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fields.Password), 2)
	usr := models.User {
		ID: uuid.New(),
		Username: fields.Username,
		Email: fields.Email,
		Password: string(hashedPassword),
		Role: models.Patient,
	}
	err = s.userClient.Persist(&usr)

	expirationTime := time.Now().Add(tokenDuration)
	token, err := utils.GenerateJWTToken(&usr, time.Now(), expirationTime)
	if err != nil {
		return "", err
	}

	to := []string{usr.Email}
	msg := []byte(
		"From: support@igaku.com\r\n" + // TODO: Read from .env
		fmt.Sprintf("To: %s\r\n", usr.Email) +
		"Subject: Igaku registration\r\n" +
		"\r\n" +
		fmt.Sprintf("Welcome %s\r\n", usr.Username),
	)

	s.mailClient.SendMail(to, msg)

	return token, nil
}
