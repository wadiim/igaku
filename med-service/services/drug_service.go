package services

import (
	"math"
	"sort"

	commonsDtos "igaku/commons/dtos"
	commonsModels "igaku/commons/models"
	commonsUtils "igaku/commons/utils"
	"igaku/med-service/dtos"
	"igaku/med-service/utils"
)

type DrugService interface {
	GetRecommendedDrugs(
		diseaseID string,
		page int, pageSize int,
		orderBy commonsModels.DrugOrderableField,
		orderMethod commonsUtils.Ordering,
	) (*commonsDtos.PaginatedResponse, error)
}

type drugService struct {
	api utils.RxClassAPI
}

func NewDrugService(api utils.RxClassAPI) DrugService {
	return &drugService{api}
}

func (s *drugService) GetRecommendedDrugs(
	diseaseID string,
	page, pageSize int,
	orderBy commonsModels.DrugOrderableField,
	orderMethod commonsUtils.Ordering,
) (*commonsDtos.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	substances, err := s.api.GetSubstances(diseaseID)
	if err != nil {
		return nil, err
	}

	var drugs []commonsModels.Drug
	for _, sub := range substances {
		d, err := s.api.GetDrugsBySubstance(sub)
		if err != nil {
			continue
		}
		drugs = append(drugs, d...)
	}

	if orderMethod == commonsUtils.Desc {
		sort.Slice(drugs, func(i, j int) bool {
			switch orderBy {
			case commonsModels.DrugID:
				return drugs[i].ID > drugs[j].ID
			case commonsModels.DrugName:
				return drugs[i].Name > drugs[j].Name
			case commonsModels.SubstanceName:
				return drugs[i].Substance > drugs[j].Substance
			default:
				return drugs[i].Name > drugs[j].Name // fallback
			}
		})
	} else {
		sort.Slice(drugs, func(i, j int) bool {
			switch orderBy {
			case commonsModels.DrugID:
				return drugs[i].ID < drugs[j].ID
			case commonsModels.DrugName:
				return drugs[i].Name < drugs[j].Name
			case commonsModels.SubstanceName:
				return drugs[i].Substance < drugs[j].Substance
			default:
				return drugs[i].Name < drugs[j].Name
			}
		})
	}

	totalCount := int64(len(drugs))
	start := (page - 1) * pageSize
	if start > len(drugs) {
		start = len(drugs)
	}
	end := start + pageSize
	if end > len(drugs) {
		end = len(drugs)
	}
	paged := drugs[start:end]

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	drugsDetails := make([]dtos.DrugDetails, len(drugs))
	for i, d := range paged {
		drugsDetails[i] = dtos.DrugDetails{
			ID: d.ID,
			Name: d.Name,
			Substance: d.Substance,
		}
	}

	resp := &commonsDtos.PaginatedResponse{
		Data: drugsDetails,
		Page: page,
		PageSize: pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}

	return resp, nil
}
