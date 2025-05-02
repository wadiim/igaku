package services

import (
	"github.com/google/uuid"

	"igaku/api-service/models"
	"igaku/api-service/repositories"
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
		return nil, err
	}

	return org, nil
}
