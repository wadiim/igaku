package services

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"fmt"
	"time"
	"errors"

	"igaku/auth-service/clients"
	"igaku/auth-service/dtos"
	"igaku/commons/models"
	"igaku/commons/utils"
	igakuErrors "igaku/auth-service/errors"
)

type AuthService interface {
	Login(creds dtos.LoginCredentials) (string, error)
	Register(fields dtos.RegistrationFields) (string, error)
}

type authService struct {
	userClient clients.UserClient
	mailClient clients.MailClient
	tokenDuration time.Duration
	from string
}

func NewAuthService(
	userClient clients.UserClient,
	mailClient clients.MailClient,
	tokenDurationInHours int,
	from string,
) (AuthService, error) {
	tokenDuration, err := time.ParseDuration(
		fmt.Sprintf("%dh", tokenDurationInHours),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to parse token duration: %w", err,
		)
	}
	return &authService{
		userClient: userClient,
		mailClient: mailClient,
		tokenDuration: tokenDuration,
		from: from,
	}, nil
}

func (s *authService) Login(creds dtos.LoginCredentials) (string, error) {
	usr, err := s.userClient.FindByUsername(creds.Username)

	if err != nil {
		if errors.Is(err, &igakuErrors.InternalError{}) {
			return "", err
		} else {
			return "", &igakuErrors.InvalidUsernameOrPasswordError{}
		}
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.Password),
		[]byte(creds.Password),
	)
	if err != nil {
		return "", &igakuErrors.InvalidUsernameOrPasswordError{}
	}

	expirationTime := time.Now().Add(s.tokenDuration)
	return utils.GenerateJWTToken(usr, time.Now(), expirationTime)
}

// TODO: Check for email duplication.
func (s *authService) Register(fields dtos.RegistrationFields) (string, error) {
	// TODO: Consider creating a UserRepository's method designed
	// specifically to check if a user with the given Username exists,
	// without fetching the user.
	_, err := s.userClient.FindByUsername(fields.Username)

	if err == nil {
		if errors.Is(err, &igakuErrors.InternalError{}) {
			return "", err
		} else {
			return "", &igakuErrors.UsernameAlreadyTakenError{}
		}
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

	expirationTime := time.Now().Add(s.tokenDuration)
	token, err := utils.GenerateJWTToken(&usr, time.Now(), expirationTime)
	if err != nil {
		return "", err
	}

	to := []string{usr.Email}
	msg := []byte(
		fmt.Sprintf("From: %s\r\n", s.from) +
		fmt.Sprintf("To: %s\r\n", usr.Email) +
		"Subject: Igaku registration\r\n" +
		"\r\n" +
		fmt.Sprintf("Welcome %s\r\n", usr.Username),
	)

	s.mailClient.SendMail(to, msg)

	return token, nil
}
