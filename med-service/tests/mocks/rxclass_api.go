package mocks

import (
	"github.com/stretchr/testify/mock"

	commonsDtos "igaku/commons/dtos"
	commonsModels "igaku/commons/models"
	commonsUtils "igaku/commons/utils"
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

func (m *MockRxClassAPI) GetDrugsBySubstance(
	substance commonsModels.Substance,
) ([]commonsModels.Drug, error) {
	args := m.Called(substance)

	var r0 []commonsModels.Drug
	if args.Get(0) != nil {
		r0 = args.Get(0).([]commonsModels.Drug)
	}
	return r0, args.Error(1)
}

func (m *MockRxClassAPI) GetRecommendedDrugs(
	diseaseID string,
	page int,
	pageSize int,
	orderBy commonsModels.DrugOrderableField,
	orderMethod commonsUtils.Ordering,
) (*commonsDtos.PaginatedResponse, error) {
	args := m.Called(diseaseID, page, pageSize, orderBy, orderMethod)

	var r0 *commonsDtos.PaginatedResponse
	if args.Get(0) != nil {
		r0 = args.Get(0).(*commonsDtos.PaginatedResponse)
	}
	return r0, args.Error(1)
}
