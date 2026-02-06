package errors

type TimeoutError struct{}

func (e *TimeoutError) Error() string {
	return "Request timed out"
}
