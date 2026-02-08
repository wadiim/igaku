package servers

import (
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	commonsErrors "igaku/commons/errors"
	commonsDtos "igaku/commons/dtos"
	"igaku/commons/models"
	"igaku/user-service/dtos"
	"igaku/user-service/services"
)

type RabbitMQServer struct {
	conn	*amqp.Connection
	ch	*amqp.Channel
	service	services.AccountService
}

func NewRabbitMQServer(
	amqpURI string,
	service services.AccountService,
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
	err := s.StartGetByIDListener()
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start `GetByIDListener`: %w",
			err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.StartFindByUsernameListener()
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start `FindByUsernameListener`: %w",
			err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.StartPersistListener()
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start `PersistListener`: %w",
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

func (s *RabbitMQServer) StartGetByIDListener() error {
	queueName := "get_by_id"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to declare a queue '%s': %w",
			queueName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to set QoS: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var resp commonsDtos.RPCResponse
			var user *dtos.AccountDetails
			var userBytes []byte

			id, err := uuid.FromBytes(d.Body)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "INVALID_ID",
					Message: err.Error(),
				}
				goto send_response
			}
			log.Printf(
				"Received RPC request for username: %s, ID: %s",
				id, d.CorrelationId,
			)

			user, err = s.service.GetAccountDetails(id)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "NOT_FOUND",
					Message: err.Error(),
				}
				goto send_response
			}

			userBytes, err = json.Marshal(user)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
				goto send_response
			}

			resp.Data = userBytes

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
			}

			publishCtx, cancelPublish := context.WithTimeout(
				context.Background(), 8*time.Second,
			)

			err = s.ch.PublishWithContext(publishCtx,
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "text/json",
					CorrelationId: d.CorrelationId,
					Body:          respBytes,
				})
			cancelPublish()

			if err != nil {
				log.Printf(
					"Failed to publish reply for ID %s: %v",
					d.CorrelationId, err,
				)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (s *RabbitMQServer) StartFindByUsernameListener() error {
	queueName := "find_by_username"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to declare a queue '%s': %w",
			queueName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to set QoS: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			username := string(d.Body)
			log.Printf(
				"Received RPC request for username: %s, ID: %s",
				username, d.CorrelationId,
			)

			var resp commonsDtos.RPCResponse
			var userBytes []byte

			user, err := s.service.GetAccountByUsername(username)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "NOT_FOUND",
					Message: err.Error(),
				}
				goto send_response
			}

			userBytes, err = json.Marshal(user)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
				goto send_response
			}

			resp.Data = userBytes

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				resp.Error = &commonsDtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
			}

			publishCtx, cancelPublish := context.WithTimeout(
				context.Background(), 8*time.Second,
			)

			err = s.ch.PublishWithContext(publishCtx,
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "text/json",
					CorrelationId: d.CorrelationId,
					Body:          respBytes,
				})
			cancelPublish()

			if err != nil {
				log.Printf(
					"Failed to publish reply for ID %s: %v",
					d.CorrelationId, err,
				)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (s *RabbitMQServer) StartPersistListener() error {
	queueName := "persist"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to declare a queue '%s': %w",
			queueName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to set QoS: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %w", err)
		return &commonsErrors.MessageBrokerError{}
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var user models.User
			var resp commonsDtos.RPCResponse

			if err := json.Unmarshal(d.Body, &user); err != nil {
				log.Printf(
					"Failed to unmarshal an RPC request: %w",
					err,
				)
				resp.Error = &commonsDtos.RPCError{
					Code: "INVALID_REQUEST",
					Message: err.Error(),
				}
				goto send_response
			}

			if err = s.service.Persist(&user); err != nil {
				code := "DATABASE_ERROR"

				var duplicatedEmailErr *commonsErrors.EmailAlreadyTakenError
				if errors.As(err, &duplicatedEmailErr) {
					code = "DUPLICATED_EMAIL"
				}

				resp.Error = &commonsDtos.RPCError{
					Code:    code,
					Message: err.Error(),
				}
				goto send_response
			}

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf(
					"Failed to marshal an RPC response: %w",
					err,
				)
				resp.Error = &commonsDtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
			}

			publishCtx, cancelPublish := context.WithTimeout(
				context.Background(), 8*time.Second,
			)

			err = s.ch.PublishWithContext(publishCtx,
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "text/json",
					CorrelationId: d.CorrelationId,
					Body:          respBytes,
				})
			cancelPublish()

			if err != nil {
				log.Printf(
					"Failed to publish reply for ID %s: %v",
					d.CorrelationId, err,
				)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}
