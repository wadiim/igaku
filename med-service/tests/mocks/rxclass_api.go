package mocks

import (
	"github.com/stretchr/testify/mock"

	"igaku/med-service/models"
)

type MockRxClassAPI struct {
	mock.Mock
}

func (m *MockRxClassAPI) GetSubstances(diseaseID string) ([]models.Substance, error) {
	args := m.Called(diseaseID)

	var r0 []models.Substance
	if args.Get(0) != nil {
		r0 = args.Get(0).([]models.Substance)
	}
	return r0, args.Error(1)
}

func (m *MockRxClassAPI) GetDrugsByName(
	name string,
) ([]models.Drug, error) {
	args := m.Called(name)

	var r0 []models.Drug
	if args.Get(0) != nil {
		r0 = args.Get(0).([]models.Drug)
	}
	return r0, args.Error(1)
}
