package services

import (
	"github.com/google/uuid"

	"igaku/dtos"
	"igaku/repositories"
)

type AccountService interface {
	GetAccountDetails(id uuid.UUID) (*dtos.AccountDetails, error)
}

type accountService struct {
       repo repositories.UserRepository
}

func NewAccountService(repo repositories.UserRepository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) GetAccountDetails(id uuid.UUID) (*dtos.AccountDetails, error) {
	user, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	details := dtos.AccountDetails{
		Username: user.Username,
		Role: string(user.Role),
	}

	return &details, nil
}
