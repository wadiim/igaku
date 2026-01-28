package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	commonsModels "igaku/commons/models"
	"igaku/med-service/models"
)

type MockMedRepository struct {
	mock.Mock
}

func (m *MockMedRepository) FindBySubstring(
	name string,
	offset int,
	limit int,
) ([]*models.Disease, error) {
	args := m.Called(name, offset, limit)

	var r0 []*models.Disease
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*models.Disease)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) AddDisease(
	rxNormID string,
	name string,
) (*models.Disease, error) {
	args := m.Called(rxNormID, name)

	var r0 *models.Disease
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Disease)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) GetDiseaseByRxNormID(
	rxNormID string,
) (*models.Disease, error) {
	args := m.Called(rxNormID)

	var r0 *models.Disease
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Disease)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) AddSubstance(
	rxClassID string,
	name string,
	subType string,
) (*models.Substance, error) {
	args := m.Called(rxClassID, name, subType)

	var r0 *models.Substance
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Substance)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) GetSubstanceByName(
	name string,
) (*models.Substance, error) {
	args := m.Called(name)

	var r0 *models.Substance
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Substance)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) AddDrug(
	rxNormID string,
	name string,
	substance string,
) (*models.Drug, error) {
	args := m.Called(rxNormID, name, substance)

	var r0 *models.Drug
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Drug)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) GetDrugByRXCUI(
	rxcui string,
) (*models.Drug, error) {
	args := m.Called(rxcui)

	var r0 *models.Drug
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Drug)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) AddMedicalHistoryItem(
	patientID uuid.UUID,
	doctorID uuid.UUID,
	drugs []models.Drug,
) (*models.MedicalHistoryItem, error) {
	args := m.Called(patientID, doctorID, drugs)

	var r0 *models.MedicalHistoryItem
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.MedicalHistoryItem)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) GetMedicalHistoryItemByPatientID(
	patientID uuid.UUID,
) ([]*models.MedicalHistoryItem, error) {
	args := m.Called(patientID)

	var r0 []*models.MedicalHistoryItem
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*models.MedicalHistoryItem)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) CountBySubstring(name string) (int64, error) {
	args := m.Called(name)

	var r0 int64
	if args.Get(0) != nil {
		r0 = args.Get(0).(int64)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) AddPatient(record *commonsModels.PatientRecord) error {
	args := m.Called(record)

	return args.Error(0)
}

func (m *MockMedRepository) FindByID(id uuid.UUID) (*commonsModels.PatientRecord, error) {
	args := m.Called(id)

	var r0 *commonsModels.PatientRecord
	if args.Get(0) != nil {
		r0 = args.Get(0).(*commonsModels.PatientRecord)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) FindByNationalID(nationalID string) (*commonsModels.PatientRecord, error) {
	args := m.Called(nationalID)

	var r0 *commonsModels.PatientRecord
	if args.Get(0) != nil {
		r0 = args.Get(0).(*commonsModels.PatientRecord)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) ValidateUniquePatient(record *commonsModels.PatientRecord) error{
	args := m.Called(record)

	return args.Error(1)
}
