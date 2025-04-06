package services

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"errors"

	"igaku/models"
	"igaku/repositories"
)

type OrganizationService interface {
	GetOrganizationByID(id uuid.UUID) (*models.Organization, error)
}

type organizationService struct {
	repo repositories.OrganizationRepository
}

func NewOrganizationService(repo repositories.OrganizationRepository) OrganizationService {
	return &organizationService{repo: repo}
}

func (s *organizationService) GetOrganizationByID(id uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.FindByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// TODO: Wrap into custom error.
			return nil, err
		}
		return nil, err
	}

	return org, nil
}
