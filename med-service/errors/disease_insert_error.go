package errors

type DiseaseInsertError struct{}

func (m *DiseaseInsertError) Error() string {
	return "Disease could not be inserted"
}
