package services

import (
	"github.com/google/uuid"

	"math"
	"sort"

	commonsDtos "igaku/commons/dtos"
	commonsModels "igaku/commons/models"
	commonsUtils "igaku/commons/utils"
	"igaku/med-service/clients"
	"igaku/med-service/dtos"
	"igaku/med-service/errors"
	"igaku/med-service/models"
	"igaku/med-service/repositories"
	"igaku/med-service/utils"
)

type MedService interface {
	GetBySubstring(name string, offset int, limit int) (*commonsDtos.PaginatedResponse, error)
	GetRecommendedDrugs(
		diseaseID string,
		page int, pageSize int,
		orderBy models.DrugOrderableField,
		orderMethod commonsUtils.Ordering,
	) (*commonsDtos.PaginatedResponse, error)
	GetDrugsByName(
		name string,
		page int, pageSize int,
		orderBy models.DrugOrderableField,
		orderMethod commonsUtils.Ordering,
	) (*commonsDtos.PaginatedResponse, error)
	ValidateUniquePatient(record *commonsModels.PatientRecord) error
	CreatePatient(data *commonsModels.PatientRecord) error
	GetPatientByNationalID(nationalID string) (*dtos.PatientDetails, error)
	AddMedicalHistoryItem(patientID uuid.UUID, doctorID uuid.UUID) error
	GetMedicalHistoryItemByPatientID(patientID uuid.UUID) (*models.MedicalHistoryItem, error)
}

type medService struct {
	api utils.RxClassAPI
	userClient clients.UserClient
	repo repositories.MedRepository
}

func NewMedService(
	api utils.RxClassAPI,
	userClient clients.UserClient,
	repo repositories.MedRepository,
) MedService {
	return &medService{api: api, userClient: userClient, repo: repo}
}

func (s *medService) GetBySubstring(
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

func (s *medService) sortDrugs(
	drugs []models.Drug,
	orderBy models.DrugOrderableField,
	orderMethod commonsUtils.Ordering,
) []models.Drug {
	if orderMethod == commonsUtils.Desc {
		sort.Slice(drugs, func(i, j int) bool {
			switch orderBy {
			case models.DrugID:
				return drugs[i].RxNormID > drugs[j].RxNormID
			case models.DrugName:
				return drugs[i].Name > drugs[j].Name
			case models.SubstanceName:
				return drugs[i].Substance > drugs[j].Substance
			default:
				return drugs[i].Name > drugs[j].Name // fallback
			}
		})
	} else {
		sort.Slice(drugs, func(i, j int) bool {
			switch orderBy {
			case models.DrugID:
				return drugs[i].RxNormID < drugs[j].RxNormID
			case models.DrugName:
				return drugs[i].Name < drugs[j].Name
			case models.SubstanceName:
				return drugs[i].Substance < drugs[j].Substance
			default:
				return drugs[i].Name < drugs[j].Name
			}
		})
	}

	return drugs
}

func (s *medService) marshalPaginatedResponse(
	drugs []models.Drug,
	page int,
	pageSize int,
) *commonsDtos.PaginatedResponse {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
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

	drugsDetailsLen := pageSize
	if len(paged) < pageSize {
		drugsDetailsLen = len(paged)
	}
	drugsDetails := make([]dtos.DrugDetails, drugsDetailsLen)
	for i, d := range paged {
		drugsDetails[i] = dtos.DrugDetails{
			ID: d.RxNormID,
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

	return resp
}

func (s *medService) GetRecommendedDrugs(
	diseaseID string,
	page, pageSize int,
	orderBy models.DrugOrderableField,
	orderMethod commonsUtils.Ordering,
) (*commonsDtos.PaginatedResponse, error) {
	substances, err := s.api.GetSubstances(diseaseID)
	if err != nil {
		return nil, err
	}

	var drugs []models.Drug
	for _, sub := range substances {
		d, err := s.api.GetDrugsByName(sub.Name)
		if err != nil {
			continue
		}
		drugs = append(drugs, d...)
	}
	if len(drugs) == 0 {
		return nil, &errors.DrugNotFoundError{}
	}

	drugs = s.sortDrugs(drugs, orderBy, orderMethod)
	resp := s.marshalPaginatedResponse(drugs, page, pageSize)

	return resp, nil
}

func (s *medService) GetDrugsByName(
	name string,
	page int, pageSize int,
	orderBy models.DrugOrderableField,
	orderMethod commonsUtils.Ordering,
) (*commonsDtos.PaginatedResponse, error) {
	drugs, err := s.api.GetDrugsByName(name)
	if err != nil {
		return nil, err
	}
	drugs = s.sortDrugs(drugs, orderBy, orderMethod)
	resp := s.marshalPaginatedResponse(drugs, page, pageSize)

	return resp, nil
}

func (s *medService) GetPatientByNationalID(nationalID string) (*dtos.PatientDetails, error) {
	record, err := s.repo.FindByNationalID(nationalID)
	if err != nil {
		return nil, err
	}

	patient, err := s.userClient.FindByID(record.ID)
	if err != nil {
		return nil, err
	}

	patientDetails := &dtos.PatientDetails{
		ID: record.ID,
		Username: patient.Username,
		Email: patient.Email,
		NationalID: record.NationalID,
	}

	return patientDetails, nil
}

func (s *medService) ValidateUniquePatient(record *commonsModels.PatientRecord) error {
	err := s.repo.ValidateUniquePatient(record)

	return err
}

func (s *medService) CreatePatient(data *commonsModels.PatientRecord) error {
	err := s.repo.AddPatient(data)

	return err
}

func (s *medService) AddMedicalHistoryItem(patientID uuid.UUID, doctorID uuid.UUID) error {
	_, err := s.repo.AddMedicalHistoryItem(patientID, doctorID)

	return err
}

func (s *medService) GetMedicalHistoryItemByPatientID(
	patientID uuid.UUID,
) (*models.MedicalHistoryItem, error) {
	item, err := s.repo.GetMedicalHistoryItemByPatientID(patientID)

	return item, err
}
