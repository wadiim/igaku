package errors

type InvalidUsernameOrPasswordError struct{}

func (m *InvalidUsernameOrPasswordError) Error() string {
	return "Invalid username or password"
}
