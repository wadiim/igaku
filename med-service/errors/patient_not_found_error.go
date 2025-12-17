package errors

type PatientNotFoundError struct{}

func (m *PatientNotFoundError) Error() string {
	return "Patient not found"
}

