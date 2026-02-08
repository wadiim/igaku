package errors

type SubstanceInsertError struct{}

func (m *SubstanceInsertError) Error() string {
	return "Substance could not be inserted"
}
