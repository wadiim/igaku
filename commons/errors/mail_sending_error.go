package errors

type MailSendingError struct {
	Err error
}

func (m *MailSendingError) Error() string {
	return "Failed to send mail"
}
