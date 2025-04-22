package services

import (
	"github.com/google/uuid"

	"math"

	"igaku/dtos"
	"igaku/models"
	"igaku/repositories"
	"igaku/utils"
)

type AccountService interface {
	GetAccountDetails(id uuid.UUID) (*dtos.AccountDetails, error)
	ListAccounts(
		page, pageSize int,
		orderBy models.UserOrderableField,
		orderMethod utils.Ordering,
	) (*dtos.PaginatedResponse, error)
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

func (s *accountService) ListAccounts(
	page, pageSize int,
	orderBy models.UserOrderableField,
	orderMethod utils.Ordering,
) (*dtos.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}

	offset := (page - 1) * pageSize

	totalCount, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}

	users, err := s.repo.FindAll(offset, pageSize, orderBy, orderMethod)
	if err != nil {
		return nil, err
	}

	accountDetailsList := make([]dtos.AccountDetailsWithID, 0, len(users))
	for _, user := range users {
		accountDetailsList = append(accountDetailsList, dtos.AccountDetailsWithID{
			ID:        user.ID.String(),
			Username:  user.Username,
			Role:      string(user.Role),
		})
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	paginatedResponse := &dtos.PaginatedResponse{
		Data:       accountDetailsList,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}

	return paginatedResponse, nil
}
