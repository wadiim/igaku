package services

import (
	"fmt"
	"net/smtp"
	"os"
)

type MailService interface {
	SendMail(to []string, msg []byte) error
}

type mailService struct {
	host string
	port string
	auth smtp.Auth
}

func NewMailService() MailService {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		host,
	)
	return &mailService{host: host, port: port, auth: auth}
}

func (s *mailService) SendMail(to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, s.auth, "api", to, msg)
	return err
}
