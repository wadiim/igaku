package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"igaku/commons/models"
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

func (m *MockMedRepository) CountBySubstring(name string) (int64, error) {
	args := m.Called(name)

	var r0 int64
	if args.Get(0) != nil {
		r0 = args.Get(0).(int64)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) AddPatient(record *models.PatientRecord) error {
	args := m.Called(record)

	return args.Error(0)
}

func (m *MockMedRepository) FindByID(id uuid.UUID) (*models.PatientRecord, error) {
	args := m.Called(id)

	var r0 *models.PatientRecord
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.PatientRecord)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) FindByNationalID(nationalID string) (*models.PatientRecord, error) {
	args := m.Called(nationalID)

	var r0 *models.PatientRecord
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.PatientRecord)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockMedRepository) ValidateUniquePatient(record *models.PatientRecord) error{
	args := m.Called(record)

	return args.Error(1)
}
