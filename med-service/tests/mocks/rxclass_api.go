package mocks

import (
	"github.com/stretchr/testify/mock"

	commonsModels "igaku/commons/models"
)

type MockRxClassAPI struct {
	mock.Mock
}

func (m *MockRxClassAPI) GetSubstances(diseaseID string) ([]commonsModels.Substance, error) {
	args := m.Called(diseaseID)

	var r0 []commonsModels.Substance
	if args.Get(0) != nil {
		r0 = args.Get(0).([]commonsModels.Substance)
	}
	return r0, args.Error(1)
}

func (m *MockRxClassAPI) GetDrugsByName(
	name string,
) ([]commonsModels.Drug, error) {
	args := m.Called(name)

	var r0 []commonsModels.Drug
	if args.Get(0) != nil {
		r0 = args.Get(0).([]commonsModels.Drug)
	}
	return r0, args.Error(1)
}
