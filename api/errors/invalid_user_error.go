package errors

import "fmt"

type InvalidUserError struct {
	Message string
}

func (m *InvalidUserError) Error() string {
	return fmt.Sprintf("Invalid user: %s", m.Message)
}
