package servers

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"

	"context"
	"errors"
	"log"
	"time"

	commonsErrors "igaku/commons/errors"
	"igaku/commons/dtos"
	"igaku/commons/models"
	"igaku/med-service/services"
)

type RabbitMQServer struct {
	conn	*amqp.Connection
	ch	*amqp.Channel
	service	services.PatientService
}

func NewRabbitMQServer(
	amqpURI string,
	service services.PatientService,
) (*RabbitMQServer, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to connect: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Printf("[RabbitMQ] Failed to open a channel: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	return &RabbitMQServer{conn: conn, ch: ch, service: service}, nil
}

func (s *RabbitMQServer) Start() error {
	err := s.StartAddPatientRecordListener()
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start `AddPatientListener`: %v",
			err,
		)
		return &commonsErrors.MessageBrokerError{}
	}
	err = s.StartValidateUniquePatientListener()
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start `ValidateUniquePatientListener`: %v",
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

func (s *RabbitMQServer) StartAddPatientRecordListener() error {
	queueName := "add_patient_record"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to declare a queue '%s': %v",
			queueName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to set QoS: %v", err)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %v", err)
		return &commonsErrors.MessageBrokerError{}
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var record models.PatientRecord
			var resp dtos.RPCResponse

			if err := json.Unmarshal(d.Body, &record); err != nil {
				log.Printf(
					"Failed to unmarshal an RPC request: %v",
					err,
				)
				resp.Error = &dtos.RPCError{
					Code: "INVALID_REQUEST",
					Message: err.Error(),
				}
				goto send_response
			}

			if err = s.service.CreatePatient(&record); err != nil {
				code := "DATABASE_ERROR"

				var duplicatedIDErr *commonsErrors.DuplicatedIDError
				var duplicatedNationalIDErr *commonsErrors.DuplicatedNationalIDError
				var invalidNationalIDErr *commonsErrors.InvalidNationalIDError
				if errors.As(err, &duplicatedIDErr) {
					code = "DUPLICATED_PATIENT_ID"
				} else if errors.As(err, &duplicatedNationalIDErr) {
					code = "DUPLICATED_PATIENT_NATIONAL_ID"
				} else if errors.As(err, &invalidNationalIDErr) {
					code = "INVALID_PATIENT_NATIONAL_ID"
				}

				resp.Error = &dtos.RPCError{
					Code:    code,
					Message: err.Error(),
				}
				goto send_response
			}

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf(
					"Failed to marshal an RPC response: %v",
					err,
				)
				resp.Error = &dtos.RPCError{
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

func (s *RabbitMQServer) StartValidateUniquePatientListener() error {
	queueName := "validate_unique_patient"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to declare a queue '%s': %v",
			queueName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	err = s.ch.Qos(1, 0, false)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to set QoS: %v", err)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %v", err)
		return &commonsErrors.MessageBrokerError{}
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var record models.PatientRecord
			var resp dtos.RPCResponse

			if err := json.Unmarshal(d.Body, &record); err != nil {
				log.Printf(
					"Failed to unmarshal an RPC request: %v",
					err,
				)
				resp.Error = &dtos.RPCError{
					Code: "INVALID_REQUEST",
					Message: err.Error(),
				}
				goto send_response
			}

			if err = s.service.ValidateUniquePatient(&record); err != nil {
				code := "DATABASE_ERROR"

				var duplicatedIDErr *commonsErrors.DuplicatedIDError
				var duplicatedNationalIDErr *commonsErrors.DuplicatedNationalIDError
				if errors.As(err, &duplicatedIDErr) {
					code = "DUPLICATED_PATIENT_ID"
				} else if errors.As(err, &duplicatedNationalIDErr) {
					code = "DUPLICATED_PATIENT_NATIONAL_ID"
				}

				resp.Error = &dtos.RPCError{
					Code:    code,
					Message: err.Error(),
				}
				goto send_response
			}

		send_response:
			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf(
					"Failed to marshal an RPC response: %v",
					err,
				)
				resp.Error = &dtos.RPCError{
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
