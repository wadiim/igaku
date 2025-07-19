package errors

type MessageBrokerError struct{}

func (m *MessageBrokerError) Error() string {
	return "Message broker failed"
}
