package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"igaku/commons/models"
)

type MockPatientRepository struct {
	mock.Mock
}

func (m *MockPatientRepository) AddPatient(record *models.PatientRecord) error {
	args := m.Called(record)

	return args.Error(0)
}

func (m *MockPatientRepository) FindByID(id uuid.UUID) (*models.PatientRecord, error) {
	args := m.Called(id)

	var r0 *models.PatientRecord
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.PatientRecord)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockPatientRepository) FindByNationalID(nationalID string) (*models.PatientRecord, error) {
	args := m.Called(nationalID)

	var r0 *models.PatientRecord
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.PatientRecord)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockPatientRepository) ValidateUniquePatient(record *models.PatientRecord) error{
	args := m.Called(record)

	return args.Error(1)
}
