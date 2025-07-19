package errors

import (
	"fmt"
)

type UsernameAlreadyTakenError struct {
	Username string
}

func (m *UsernameAlreadyTakenError) Error() string {
	return fmt.Sprintf("Username '%s' already taken", m.Username)
}

