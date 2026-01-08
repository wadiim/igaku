package errors

type RxClassUnavailableError struct {}

func (m *RxClassUnavailableError) Error() string {
	return "RxClass API Unavailable"
}
