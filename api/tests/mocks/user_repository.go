package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"igaku/models"
	"igaku/utils"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)

	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *UserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)

	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *UserRepository) FindAll(
	offset, limit int,
	orderBy models.UserOrderableField, orderMethod utils.Ordering,
) ([]models.User, error) {
	args := m.Called(offset, limit, orderBy, orderMethod)

	var r0 []models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).([]models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *UserRepository) CountAll() (int64, error) {
	args := m.Called()

	return args.Get(0).(int64), args.Error(1)
}
