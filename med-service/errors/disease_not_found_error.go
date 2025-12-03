package errors

type DiseaseNotFoundError struct{}

func (m *DiseaseNotFoundError) Error() string {
	return "Disease not found"
}

