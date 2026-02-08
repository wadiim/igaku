package errors

type OffsetNegativeError struct{}

func (m *OffsetNegativeError) Error() string {
	return "Offset must be positive"
}
