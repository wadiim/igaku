package errors

type TokenGenerationError struct{}

func (m *TokenGenerationError) Error() string {
	return "Failed to generate token"
}
