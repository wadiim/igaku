package servers

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"

	"log"

	"igaku/commons/dtos"
	"igaku/mail-service/services"
	commonsErrors "igaku/commons/errors"
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
		log.Printf("[RabbitMQ] Failed to connect: %w", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Printf("[RabbitMQ] Failed to open a channel: %w", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	return &RabbitMQServer{conn: conn, ch: ch, service: service}, nil
}

func (s *RabbitMQServer) Start() error {
	err := s.StartSendMailListener()
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start `SendMailListener`: %w",
			err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

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
		log.Printf(
			"[RabbitMQ] Failed to declare an exchange '%s': %w",
			exchangeName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	q, err := s.ch.QueueDeclare(
		"", false, false, true, false, nil,
	)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to declare a queue: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.ch.QueueBind(
		q.Name, "", exchangeName, false, nil,
	)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to bind queue '%s': %w",
			q.Name, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %w", err)
		return &commonsErrors.MessageBrokerError{}
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
				continue
			}
		}
	}()

	return nil
}
