package errors

type MedicalHistoryItemNotFoundError struct{}

func (m *MedicalHistoryItemNotFoundError) Error() string {
	return "Medical history item not found"
}
