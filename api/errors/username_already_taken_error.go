package errors

type UsernameAlreadyTakenError struct{}

func (m *UsernameAlreadyTakenError) Error() string {
	return "Username already taken"
}

