package errors

import (
	"fmt"
)

type EmailAlreadyTakenError struct {
	Email string
}

func (m *EmailAlreadyTakenError) Error() string {
	return fmt.Sprintf("Email '%s' already taken", m.Email)
}

