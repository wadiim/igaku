package errors

type MedicalHistoryItemInsertError struct{}

func (m *MedicalHistoryItemInsertError) Error() string {
	return "Medical history item could not be inserted"
}
