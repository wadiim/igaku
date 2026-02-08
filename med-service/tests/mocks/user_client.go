package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"igaku/commons/models"
)

type UserClient struct {
	mock.Mock
}

func (m *UserClient) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)

	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *UserClient) Shutdown() {}
