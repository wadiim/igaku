package mocks

import (
	"github.com/stretchr/testify/mock"

	"igaku/commons/models"
)

type UserClient struct {
	mock.Mock
}

func (m *UserClient) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)

	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *UserClient) Persist(user *models.User) error {
	args := m.Called(user)

	return args.Error(0)
}
