package mocks

import (
	"github.com/stretchr/testify/mock"
	"igaku/commons/dtos"
)

type GeoService struct {
	mock.Mock
}

func (m *GeoService) Search(
	address string,
) ([]dtos.Location, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).([]dtos.Location), nil
}

func (m *GeoService) Reverse(
	lat, lon string,
) (*dtos.Location, error) {
	args := m.Called(lat, lon)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*dtos.Location), nil
}
