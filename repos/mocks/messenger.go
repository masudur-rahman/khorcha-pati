package mocks

import (
	"github.com/masudur-rahman/khorcha-pati/modules/auth"

	"github.com/stretchr/testify/mock"
)

// Messenger is a mock for auth.Messenger.
type Messenger struct {
	mock.Mock
}

var _ auth.Messenger = &Messenger{}

func (m *Messenger) SendMessage(telegramID int64, text string) error {
	return m.Called(telegramID, text).Error(0)
}
