package errors

type InvalidAddressError struct {}

func (m *InvalidAddressError) Error() string {
	return "Invalid address"
}


