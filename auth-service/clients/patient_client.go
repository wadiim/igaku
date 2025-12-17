package clients

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"igaku/auth-service/errors"
	"igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
	"igaku/commons/models"
)

type PatientClient interface {
	AddPatientRecord(record *models.PatientRecord) error
	ValidateUniquePatient(record *models.PatientRecord) error
	Shutdown()
}

type patientClient struct {
	url	string
	conn	*amqp.Connection
	ch	*amqp.Channel
	replyMsgs	<-chan amqp.Delivery
	pendingCalls	sync.Map
}

func (c *patientClient) Shutdown() {
	if c.ch != nil { c.ch.Close() }
	if c.conn != nil { c.conn.Close() }
}

// type responseChan struct {
// 	ch	chan []byte
// 	err	chan error
// }

func NewPatientClient(url string) (PatientClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to connect: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[RabbitMQ] Failed to create a channel: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	replyMsgs, err := ch.Consume(
		"amq.rabbitmq.reply-to", "",
		true, true, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		log.Printf(
			"[RabbitMQ] Failed to consume `reply-to` queue: %v",
			err,
		)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	client := &patientClient{
		url: url, conn: conn, ch: ch, replyMsgs: replyMsgs,
	}

	go client.listen()

	return client, nil
}

func (c *patientClient) listen() {
	for msg := range c.replyMsgs {
		if val, ok := c.pendingCalls.Load(msg.CorrelationId); ok {
			res := val.(*responseChan)
			select {
			case res.ch <- msg.Body:
			default:
			}
			c.pendingCalls.Delete(msg.CorrelationId)
		}
	}
}

func (c *patientClient) call(routingKey string, body []byte) ([]byte, error) {
	corrID := uuid.New().String()
	res := &responseChan{
		ch:	make(chan []byte, 1),
		err:	make(chan error, 1),
	}

	c.pendingCalls.Store(corrID, res)

	err := c.ch.Publish(
		"",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:	"application/json",
			CorrelationId:	corrID,
			ReplyTo:	"amq.rabbitmq.reply-to",
			Body:		body,
		},
	)
	if err != nil {
		c.pendingCalls.Delete(corrID)
		log.Printf("[RabbitMQ] Failed to publish a message: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case reply := <-res.ch:
		return reply, nil
	case <-ctx.Done():
		c.pendingCalls.Delete(corrID)
		log.Println("[RabbitMQ] Timeout waiting for RPC response")
		return nil, &commonsErrors.MessageBrokerError{}
	}
}

func (c *patientClient) AddPatientRecord(record *models.PatientRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		log.Printf("Failed to marshal patient record: %v", err)
	}

	reply, err := c.call("add_patient_record", []byte(body))
	if err != nil {
		errmsg := fmt.Sprintf(
			"[RabbitMQ] Failed to publish request to add_patient_record: %v",
			err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}

	var rpcResp dtos.RPCResponse
	if err := json.Unmarshal(reply, &rpcResp); err != nil {
		errmsg := fmt.Sprintf(
			"[RabbitMQ] Failed to unmarshal RPC response: %v\n", err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}

	if rpcResp.Error != nil {
		if rpcResp.Error.Code == "DUPLICATED_PATIENT_ID" {
			errmsg := fmt.Sprintf(
				"Failed to add patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &commonsErrors.DuplicatedIDError{
				ID: record.ID,
			}
		} else if rpcResp.Error.Code == "DUPLICATED_PATIENT_NATIONAL_ID" {
			errmsg := fmt.Sprintf(
				"Failed to add patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &commonsErrors.DuplicatedNationalIDError{
				NationalID: record.NationalID,
			}
		} else if rpcResp.Error.Code == "INVALID_PATIENT_NATIONAL_ID" {
			errmsg := fmt.Sprintf(
				"Failed to add patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &commonsErrors.InvalidNationalIDError{
				NationalID: record.NationalID,
			}
		} else {
			errmsg := fmt.Sprintf(
				"Failed to add patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &errors.InternalError{}
		}
	}

	return nil
}

func (c *patientClient) ValidateUniquePatient(record *models.PatientRecord) error {
	body, err := json.Marshal(record)
	if err != nil {
		log.Printf("Failed to marshal a user: %v", err)
		return &errors.InternalError{}
	}

	reply, err := c.call("validate_unique_patient", []byte(body))
	if err != nil {
		errmsg := fmt.Sprintf(
			"[RabbitMQ] Failed to publish request to validate_unique_patient: %v",
			err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}

	var rpcResp dtos.RPCResponse
	if err := json.Unmarshal(reply, &rpcResp); err != nil {
		errmsg := fmt.Sprintf(
			"[RabbitMQ] Failed to unmarshal RPC response: %v\n", err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}
	
	if rpcResp.Error != nil {
		if rpcResp.Error.Code == "DUPLICATED_PATIENT_ID" {
			errmsg := fmt.Sprintf(
				"Failed to validate patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &commonsErrors.DuplicatedIDError{
				ID: record.ID,
			}
		} else if rpcResp.Error.Code == "DUPLICATED_PATIENT_NATIONAL_ID" {
			errmsg := fmt.Sprintf(
				"Failed to validate patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &commonsErrors.DuplicatedNationalIDError{
				NationalID: record.NationalID,
			}
		} else {
			errmsg := fmt.Sprintf(
				"Failed to validate patient: %v",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return &errors.InternalError{}
		}
	}

	return nil
}
