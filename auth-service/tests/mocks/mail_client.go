package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MailClient struct {
	mock.Mock
}

func (m *MailClient) SendMail(to []string, msg []byte) error {
	args := m.Called(to, msg)

	return args.Error(0)
}

func (m *MailClient) Shutdown() {}

