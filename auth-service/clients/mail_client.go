package clients

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"igaku/commons/dtos"
)

type MailClient interface {
	SendMail(to []string, msg []byte) error
	Shutdown()
}

type mailClient struct {
	url	string
	conn	*amqp.Connection
	ch	*amqp.Channel
}

func NewMailClient(url string) (MailClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to create a channel")
		return nil, err
	}

	return &mailClient{url: url, conn: conn, ch: ch}, nil
}

func (s *mailClient) Shutdown() {
	if s.ch != nil { s.ch.Close() }
	if s.conn != nil { s.conn.Close() }
}

// TODO: Use custom errors
func (c *mailClient) SendMail(to []string, msg []byte) error {
	exchangeName := "mail"

	err := c.ch.ExchangeDeclare(
		exchangeName, "fanout", true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf(
			"Failed to declare an exchange '%s': %w",
			exchangeName, err,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sendMailReq := dtos.SendMailRequest{To: to, Msg: msg}
	body, err := json.Marshal(sendMailReq)
	if err != nil {
		log.Println("Failed to marshal the `SendMailRequest`")
		return err
	}

	err = c.ch.PublishWithContext(
		ctx, exchangeName, "", false, false,
		amqp.Publishing{
			ContentType:	"text/json",
			Body:		[]byte(body),
		},
	)
	if err != nil {
		errmsg := fmt.Sprintf(
			"Failed to publish request to send mail: %w", err,
		)
		log.Println(errmsg)
		return fmt.Errorf(errmsg)
	}

	return nil
}
