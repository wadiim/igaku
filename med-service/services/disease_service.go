package services

import (
	"igaku/commons/models"
	"igaku/med-service/repositories"
)

type DiseaseService interface {
	GetByName(name string) ([]*models.Disease, error)
}

type diseaseService struct {
	repo repositories.DiseaseRepository
}

func NewDiseaseService(repo repositories.DiseaseRepository) DiseaseService {
	return &diseaseService{repo}
}

func (s *diseaseService) GetByName(name string) ([]*models.Disease, error) {
	diseases, err := s.repo.FindByName(name)

	if err != nil {
		return nil, err
	}

	return diseases, nil

}
