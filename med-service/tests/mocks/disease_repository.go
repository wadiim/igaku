package mocks

import (
	"github.com/stretchr/testify/mock"

	"igaku/commons/models"
)

type MockDiseaseRepository struct {
	mock.Mock
}

func (m *MockDiseaseRepository) FindBySubstring(
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

func (m *MockDiseaseRepository) CountBySubstring(name string) (int64, error) {
	args := m.Called(name)

	var r0 int64
	if args.Get(0) != nil {
		r0 = args.Get(0).(int64)
	}

	r1 := args.Error(1)

	return r0, r1
}
