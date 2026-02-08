package errors

type RxNormUnavailableError struct {}

func (m *RxNormUnavailableError) Error() string {
	return "RxNorm API Unavailable"
}
