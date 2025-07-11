package errors

type DuplicatedEmailError struct {
	Message string
}

func (m *DuplicatedEmailError) Error() string {
	return m.Message
}

