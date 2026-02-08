package errors

type DrugNotFoundError struct{}

func (m *DrugNotFoundError) Error() string {
	return "Drug not found"
}
