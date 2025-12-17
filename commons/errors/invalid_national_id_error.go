package errors

import (
	"fmt"
)

type InvalidNationalIDError struct {
	NationalID string
}

func (m *InvalidNationalIDError) Error() string {
	return fmt.Sprintf("Invalid ID:", m.NationalID)
}
