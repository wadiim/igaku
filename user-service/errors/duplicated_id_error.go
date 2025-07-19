package errors

import (
	"github.com/google/uuid"

	"fmt"
)

type DuplicatedIDError struct {
	ID uuid.UUID
}

func (m *DuplicatedIDError) Error() string {
	return fmt.Sprintf("Duplicated ID:", m.ID)
}

