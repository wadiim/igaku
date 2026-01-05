package services

import (
	"math"

	commonsDtos "igaku/commons/dtos"
	"igaku/med-service/dtos"
	"igaku/med-service/repositories"
)

type DiseaseService interface {
	GetBySubstring(name string, offset int, limit int) (*commonsDtos.PaginatedResponse, error)
}

type diseaseService struct {
	repo repositories.DiseaseRepository
}

func NewDiseaseService(repo repositories.DiseaseRepository) DiseaseService {
	return &diseaseService{repo}
}

func (s *diseaseService) GetBySubstring(
	name string,
	page int,
	pageSize int,
) (*commonsDtos.PaginatedResponse, error) {

	offset := (page - 1) * pageSize
	diseases, err := s.repo.FindBySubstring(name, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.repo.CountBySubstring(name)
	if err != nil {
		return nil, err
	}

	diseaseDetailsList := make([]dtos.DiseaseDetails, len(diseases))
	for i, disease := range diseases {
		diseaseDetailsList[i] = dtos.DiseaseDetails{
			ID: disease.ID.String(),
			RxNormID: disease.RxNormID,
			Name: disease.Name,
		}
	}
	
	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	paginatedResponse := &commonsDtos.PaginatedResponse{
		Data:       diseaseDetailsList,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}

	return paginatedResponse, nil
}
