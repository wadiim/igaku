package errors

type DatabaseError struct{}

func (m *DatabaseError) Error() string {
	return "Database error"
}

