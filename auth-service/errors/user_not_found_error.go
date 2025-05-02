package errors

type UserNotFoundError struct{}

func (m *UserNotFoundError) Error() string {
	return "User not found"
}

