package errors

type LimitNegativeError struct{}

func (m *LimitNegativeError) Error() string {
	return "Limit must be positive"
}
