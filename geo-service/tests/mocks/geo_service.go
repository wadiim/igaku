package mocks

import (
	"github.com/stretchr/testify/mock"
	geoDtos "igaku/geo-service/dtos"
)

type GeoService struct {
	mock.Mock
}

func (m *GeoService) Search(
	address string,
) ([]geoDtos.Location, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).([]geoDtos.Location), nil
}
