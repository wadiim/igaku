package servers

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"

	"fmt"
	"log"

	"igaku/commons/dtos"
	"igaku/mail-service/services"
)

type RabbitMQServer struct {
	conn	*amqp.Connection
	ch	*amqp.Channel
	service	services.MailService
}

func NewRabbitMQServer(
	amqpURI string,
	service services.MailService,
) (*RabbitMQServer, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to open a channel: %w", err)
	}

	return &RabbitMQServer{conn: conn, ch: ch, service: service}, nil
}

func (s *RabbitMQServer) Start() error {
	err := s.StartSendMailListener()
	if err != nil { return err }

	return nil
}

func (s *RabbitMQServer) Shutdown() {
	if s.ch != nil { s.ch.Close() }
	if s.conn != nil { s.conn.Close() }
}

func (s *RabbitMQServer) StartSendMailListener() error {
	exchangeName := "mail"

	err := s.ch.ExchangeDeclare(
		exchangeName, "fanout", true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf(
			"Failed to declare an exchange '%s': %w",
			exchangeName, err,
		)
	}

	q, err := s.ch.QueueDeclare(
		"", false, false, true, false, nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %w", err)
	}

	err = s.ch.QueueBind(
		q.Name, "", exchangeName, false, nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to bind queue '%s': %w", q.Name, err)
	}

	msgs, err := s.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Failed to register a consumer: %w", err)
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var mailReq dtos.SendMailRequest
			if err := json.Unmarshal(d.Body, &mailReq); err != nil {
				log.Printf(
					"Failed to unmarshal request: %w", err,
				)
				continue
			}

			if err = s.service.SendMail(mailReq.To, mailReq.Msg); err != nil {
				log.Printf(
					"Failed to send mail to '%s': %w",
					mailReq.To, err,
				)
			}
		}
	}()

	return nil
}
