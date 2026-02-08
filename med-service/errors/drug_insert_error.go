package errors

type DrugInsertError struct{}

func (m *DrugInsertError) Error() string {
	return "Drug could not be inserted"
}
