package errors

type ExternalApiRequestError struct {
	Message string
}

func (m *ExternalApiRequestError) Error() string {
	return m.Message
}

