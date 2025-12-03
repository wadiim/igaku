package errors

type RequestError struct {
	Message string
}

func (m *RequestError) Error() string {
	return m.Message
}

