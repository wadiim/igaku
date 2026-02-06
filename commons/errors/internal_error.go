package errors

type InternalError struct{}

func (m *InternalError) Error() string {
	return "Internal error"
}
