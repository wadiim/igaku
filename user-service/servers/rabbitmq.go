package servers

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"

	"context"
	"fmt"
	"log"
	"time"

	"igaku/commons/dtos"
	"igaku/commons/models"
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
	err := s.StartFindByUsernameListener()
	if err != nil { return err }

	err = s.StartPersistListener()
	if err != nil { return err }

	return nil
}

func (s *RabbitMQServer) Shutdown() {
	if s.ch != nil { s.ch.Close() }
	if s.conn != nil { s.conn.Close() }
}

func (s *RabbitMQServer) StartFindByUsernameListener() error {
	queueName := "find_by_username"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue '%s': %w", queueName, err)
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("Failed to set QoS: %w", err)
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Failed to register a consumer: %w", err)
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			username := string(d.Body)
			log.Printf(
				"Received RPC request for username: %s, ID: %s",
				username, d.CorrelationId,
			)

			var resp dtos.RPCResponse
			var userBytes []byte

			user, err := s.service.GetAccountByUsername(username)
			if err != nil {
				resp.Error = &dtos.RPCError{
					Code: "NOT_FOUND",
					Message: err.Error(),
				}
				goto send_response
			}

			userBytes, err = json.Marshal(user)
			if err != nil {
				resp.Error = &dtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
				goto send_response
			}

			resp.Data = userBytes

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				resp.Error = &dtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
			}

			publishCtx, cancelPublish := context.WithTimeout(
				context.Background(), 5*time.Second,
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
		return fmt.Errorf("Failed to declare a queue '%s': %w", queueName, err)
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("Failed to set QoS: %w", err)
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Failed to register a consumer: %w", err)
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var user models.User
			var resp dtos.RPCResponse

			if err := json.Unmarshal(d.Body, &user); err != nil {
				resp.Error = &dtos.RPCError{
					Code: "INVALID_REQUEST",
					Message: err.Error(),
				}
				goto send_response
			}

			if err = s.service.Persist(&user); err != nil {
				resp.Error = &dtos.RPCError{
					Code: "DATABASE_ERROR",
					Message: err.Error(),
				}
				goto send_response
			}

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				resp.Error = &dtos.RPCError{
					Code: "INTERNAL",
					Message: err.Error(),
				}
			}

			publishCtx, cancelPublish := context.WithTimeout(
				context.Background(), 5*time.Second,
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
