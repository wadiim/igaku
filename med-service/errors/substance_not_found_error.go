package errors

type SubstanceNotFoundError struct{}

func (m *SubstanceNotFoundError) Error() string {
	return "Substance not found"
}
