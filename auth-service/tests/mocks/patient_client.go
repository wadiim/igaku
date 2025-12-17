package mocks

import (
	"github.com/stretchr/testify/mock"

	"igaku/commons/models"
)

type PatientClient struct {
	mock.Mock
}

func (m *PatientClient) AddPatientRecord(record *models.PatientRecord) error {
	args := m.Called(record)

	return args.Error(0)
}

func (m *PatientClient) ValidateUniquePatient(record *models.PatientRecord) error {
	args := m.Called(record)

	return args.Error(0)
}

func (m *PatientClient) Shutdown() {}
