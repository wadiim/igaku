package errors

import (
	"fmt"
)

type DuplicatedNationalIDError struct {
	NationalID string
}

func (m *DuplicatedNationalIDError) Error() string {
	return fmt.Sprintf("Duplicated ID: %v", m.NationalID)
}
